// Package discovery shares Nacos HTTP subscriptions across gateway snapshots.
// It never owns the Naming SDK: the process closes that shared dependency.
package discovery

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	registrynacos "sitehub/common/registry/nacos"
)

type Node struct {
	URL    string
	Weight int
}

type Manager struct {
	naming   registrynacos.Naming
	group    string
	maxStale time.Duration
	mu       sync.Mutex
	closed   bool
	services map[string]*subscription
}

func New(naming registrynacos.Naming, group string, maxStale time.Duration) *Manager {
	return &Manager{naming: naming, group: group, maxStale: maxStale, services: make(map[string]*subscription)}
}

var servicePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*-api$`)

// Watch shares one SDK subscription per service. Updates are serialized for
// each watcher and delivered without internal locks; slow callbacks coalesce
// pending updates to the newest snapshot. The initial callback is asynchronous.
// cancel is idempotent. A callback already executing may finish after cancel;
// callbacks may themselves cancel their watch or close the manager safely.
func (m *Manager) Watch(service string, update func([]Node)) (cancel func(), err error) {
	if m.naming == nil || m.maxStale <= 0 || strings.TrimSpace(m.group) == "" || !servicePattern.MatchString(service) || update == nil {
		return nil, fmt.Errorf("discovery requires a naming client, group, positive MaxStale, service-api name and callback")
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, fmt.Errorf("discovery manager is closed")
	}
	s, exists := m.services[service]
	if !exists {
		s = &subscription{manager: m, ready: make(chan struct{}), done: make(chan struct{}), stopped: make(chan struct{}), watchers: make(map[*watcher]struct{}), lastGood: time.Now()}
		s.param = &vo.SubscribeParam{ServiceName: service, GroupName: m.group, SubscribeCallback: s.push}
		m.services[service] = s
	}
	// Pending Watch callers also keep the shared subscription alive while its
	// initial selection is completing or another caller cancels its watch.
	s.pending++
	m.mu.Unlock()
	if !exists {
		s.start()
	} else {
		<-s.ready
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s.pending--
	if m.closed {
		return nil, fmt.Errorf("discovery manager is closed")
	}
	if s.startErr != nil {
		if s.pending == 0 {
			delete(m.services, service)
			s.close()
		}
		return nil, fmt.Errorf("subscribe to %s failed", service)
	}
	w := &watcher{update: update, changes: make(chan []Node, 1), done: make(chan struct{})}
	s.mu.Lock()
	s.watchers[w] = struct{}{}
	w.enqueue(s.nodes)
	s.mu.Unlock()
	go w.run()
	var once sync.Once
	return func() {
		once.Do(func() {
			m.mu.Lock()
			defer m.mu.Unlock()
			s.mu.Lock()
			if _, exists := s.watchers[w]; exists {
				delete(s.watchers, w)
				close(w.done)
			}
			empty := len(s.watchers) == 0
			s.mu.Unlock()
			if !m.closed && empty && s.pending == 0 {
				delete(m.services, service)
				// Serialize unsubscribe against the next Watch for this name: an
				// old SDK unsubscribe must never cancel a replacement subscription.
				s.close()
			}
		})
	}, nil
}

// Close releases every subscription and monitoring goroutine, without closing
// the shared SDK. It does not wait for user callbacks, which may call Close.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	m.closed = true
	for _, s := range m.services {
		s.close()
	}
	m.services = make(map[string]*subscription)
}

type subscription struct {
	manager    *Manager
	param      *vo.SubscribeParam
	ready      chan struct{}
	done       chan struct{}
	stopped    chan struct{}
	startErr   error
	subscribed bool
	pending    int // manager.mu
	mu         sync.Mutex
	closed     bool
	revision   uint64
	lastGood   time.Time
	expired    bool
	nodes      []Node
	watchers   map[*watcher]struct{}
}

func (s *subscription) start() {
	s.startErr = s.manager.naming.Subscribe(s.param)
	if s.startErr == nil {
		s.subscribed = true
		s.refresh()
		go s.monitor()
	} else {
		close(s.stopped)
	}
	close(s.ready)
}

func (s *subscription) push(instances []model.Instance, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	// Errors also invalidate an in-flight selection: a response started before
	// a newer push (including a failure) must not overwrite that newer event.
	s.revision++
	if err == nil {
		s.applyLocked(instances)
	}
}

func (s *subscription) refresh() {
	if !s.manager.naming.ServerHealthy() {
		return
	}
	s.mu.Lock()
	revision, closed := s.revision, s.closed
	s.mu.Unlock()
	if closed {
		return
	}
	instances, err := s.manager.naming.SelectInstances(vo.SelectInstancesParam{ServiceName: s.param.ServiceName, GroupName: s.param.GroupName, HealthyOnly: true})
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.revision != revision || err != nil || !s.manager.naming.ServerHealthy() {
		return
	}
	s.applyLocked(instances)
}

func (s *subscription) applyLocked(instances []model.Instance) {
	s.lastGood = time.Now()
	s.expired = false
	s.publishLocked(nodesFromInstances(instances))
}

func (s *subscription) publishLocked(nodes []Node) {
	if slices.Equal(s.nodes, nodes) {
		return
	}
	s.nodes = nodes
	for w := range s.watchers {
		w.enqueue(nodes)
	}
}

func (s *subscription) monitor() {
	defer close(s.stopped)
	refreshes := make(chan struct{}, 1)
	refreshStopped := make(chan struct{})
	go func() {
		defer close(refreshStopped)
		for {
			select {
			case <-s.done:
				return
			case <-refreshes:
				s.refresh()
			}
		}
	}()
	defer func() { <-refreshStopped }()
	ticker := time.NewTicker(max(time.Nanosecond, min(time.Second, s.manager.maxStale/4)))
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			// Expiry remains independent of a slow SDK read. The SDK applies its
			// own call timeout; only one refresh may be in flight per service.
			select {
			case refreshes <- struct{}{}:
			default:
			}
			s.mu.Lock()
			if !s.closed && !s.expired && time.Since(s.lastGood) >= s.manager.maxStale {
				s.expired = true
				s.publishLocked(nil)
			}
			s.mu.Unlock()
		}
	}
}

func (s *subscription) close() {
	<-s.ready
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	close(s.done)
	for w := range s.watchers {
		close(w.done)
		delete(s.watchers, w)
	}
	s.mu.Unlock()
	if s.subscribed {
		_ = s.manager.naming.Unsubscribe(s.param)
	}
	<-s.stopped
}

type watcher struct {
	update  func([]Node)
	changes chan []Node
	done    chan struct{}
}

// Called under subscription.mu so publishers cannot race and reorder values.
func (w *watcher) enqueue(nodes []Node) {
	nodes = slices.Clone(nodes)
	select {
	case w.changes <- nodes:
	default:
		select {
		case <-w.changes:
		default:
		}
		w.changes <- nodes
	}
}

func (w *watcher) run() {
	for {
		select {
		case <-w.done:
			return
		case nodes := <-w.changes:
			select {
			case <-w.done:
				return
			default:
				w.update(nodes)
			}
		}
	}
}

func nodesFromInstances(instances []model.Instance) []Node {
	byURL := make(map[string]int, len(instances))
	for _, instance := range instances {
		ip := net.ParseIP(instance.Ip)
		if !instance.Healthy || !instance.Enable || instance.Weight <= 0 || math.IsNaN(instance.Weight) || math.IsInf(instance.Weight, 0) || instance.Port == 0 || instance.Port > 65535 || ip == nil || ip.IsUnspecified() || ip.IsMulticast() {
			continue
		}
		address := "http://" + net.JoinHostPort(ip.String(), strconv.FormatUint(instance.Port, 10))
		// Nacos permits fractional weights; integer gateway weights preserve
		// eligibility and bound load-balancer arithmetic for hostile metadata.
		weight := int(max(1, min(1_000_000, math.Round(instance.Weight))))
		byURL[address] = max(byURL[address], weight)
	}
	nodes := make([]Node, 0, len(byURL))
	for address, weight := range byURL {
		nodes = append(nodes, Node{URL: address, Weight: weight})
	}
	slices.SortFunc(nodes, func(a, b Node) int { return strings.Compare(a.URL, b.URL) })
	return nodes
}
