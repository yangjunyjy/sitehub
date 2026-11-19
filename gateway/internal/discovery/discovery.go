// Package discovery shares service subscriptions across gateway snapshots.
// The process owns and closes the provider after all discovery consumers stop.
package discovery

import (
	"context"
	"fmt"
	"math"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"sitehub/common/registry"
)

type Node struct {
	URL    string
	Weight int
}

type Manager struct {
	provider registry.Discovery
	maxStale time.Duration
	mu       sync.Mutex
	closed   bool
	services map[string]*subscription
}

func New(provider registry.Discovery, maxStale time.Duration) *Manager {
	return &Manager{provider: provider, maxStale: maxStale, services: make(map[string]*subscription)}
}

// HTTP upstreams may use any safe service identifier; gateway discovery does
// not impose the application's domain names or an "-api" naming convention.
var servicePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,127}$`)

// Watch shares one provider subscription per service. Updates are serialized for
// each watcher and delivered without internal locks; slow callbacks coalesce
// pending updates to the newest snapshot. The initial callback is asynchronous.
// cancel is idempotent. A callback already executing may finish after cancel;
// callbacks may themselves cancel their watch or close the manager safely.
func (m *Manager) Watch(service string, update func([]Node)) (cancel func(), err error) {
	if m.provider == nil || m.maxStale <= 0 || !servicePattern.MatchString(service) || update == nil {
		return nil, fmt.Errorf("discovery requires a provider, positive MaxStale, valid service name and callback")
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil, fmt.Errorf("discovery manager is closed")
	}
	s, exists := m.services[service]
	if !exists {
		ctx, cancel := context.WithCancel(context.Background())
		s = &subscription{manager: m, service: service, ctx: ctx, cancel: cancel, ready: make(chan struct{}), done: make(chan struct{}), stopped: make(chan struct{}), watchers: make(map[*watcher]struct{}), lastGood: time.Now()}
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
				// old cancellation must never cancel a replacement subscription.
				s.close()
			}
		})
	}, nil
}

// Close releases every subscription and monitoring goroutine, without closing
// the provider. It does not wait for user callbacks, which may call Close.
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
	manager  *Manager
	service  string
	ctx      context.Context
	cancel   context.CancelFunc
	unwatch  func()
	ready    chan struct{}
	done     chan struct{}
	stopped  chan struct{}
	startErr error
	pending  int // manager.mu
	mu       sync.Mutex
	closed   bool
	revision uint64
	lastGood time.Time
	expired  bool
	nodes    []Node
	watchers map[*watcher]struct{}
}

func (s *subscription) start() {
	s.unwatch, s.startErr = s.manager.provider.Watch(s.ctx, s.service, s.push)
	if s.startErr == nil {
		s.refresh()
		go s.monitor()
	} else {
		s.cancel()
		close(s.stopped)
	}
	close(s.ready)
}

func (s *subscription) push(instances []registry.Instance, err error) {
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
	if !s.manager.provider.Healthy() {
		return
	}
	s.mu.Lock()
	revision, closed := s.revision, s.closed
	s.mu.Unlock()
	if closed {
		return
	}
	instances, err := s.manager.provider.List(s.ctx, s.service)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.revision != revision || err != nil || !s.manager.provider.Healthy() {
		return
	}
	s.applyLocked(instances)
}

func (s *subscription) applyLocked(instances []registry.Instance) {
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
			// Expiry remains independent of a slow provider read. The provider applies its
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
	s.cancel()
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
	if s.unwatch != nil {
		s.unwatch()
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

func nodesFromInstances(instances []registry.Instance) []Node {
	byURL := make(map[string]int, len(instances))
	for _, instance := range instances {
		ip := net.ParseIP(instance.IP)
		if !instance.Healthy || !instance.Enabled || instance.Weight <= 0 || math.IsNaN(instance.Weight) || math.IsInf(instance.Weight, 0) || instance.Port == 0 || instance.Port > 65535 || ip == nil || ip.IsUnspecified() || ip.IsMulticast() {
			continue
		}
		address := "http://" + net.JoinHostPort(ip.String(), strconv.FormatUint(instance.Port, 10))
		// Providers may use fractional weights; integer gateway weights preserve
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
