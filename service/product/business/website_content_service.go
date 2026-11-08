package services

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"

	errorx "sitehub/common/errorx"
	settingsservice "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	filecontract "sitehub/service/file/contract"
	productcontract "sitehub/service/product/contract"
	productdomain "sitehub/service/product/domain"
	productrepository "sitehub/service/product/model"
)

const (
	websiteContentKey   = "website.content"
	websiteDownloadsKey = "website.downloads"
	websiteStorageKey   = "website.storage"
)

var ErrInvalidWebsiteDownloads = errors.New("下载内容校验失败")

type WebsiteContentService interface {
	EnsureDefaults() error
	GetContent() (productcontract.WebsiteContent, error)
	GetDownloads() (productcontract.WebsiteDownloads, error)
	GetPublicDownloads() (productcontract.WebsiteDownloads, error)
	GetPublishedDownload(id string) (productcontract.ProductDownload, error)
	GetDocs() (productcontract.WebsiteDocs, error)
	GetDocModes() (productcontract.WebsiteDocModes, error)
	ChangeDocMode(productdomain.DocModeInput) error
	GetStorage() (filecontract.WebsiteStorage, error)
	GetDocsNavigation(version string, mode ...string) (productcontract.WebsiteDocsNavigation, error)
	GetDocument(id, version string, mode ...string) (productcontract.WebsiteDocDocument, error)
	SaveContent(content productcontract.WebsiteContent) error
	SaveDownloads(downloads productcontract.WebsiteDownloads) error
	GetAdminDocument(id, version string) (productcontract.WebsiteDocDocument, error)
	SaveDocSettings(productdomain.DocSettingsInput) error
	ChangeDocVersion(productdomain.DocVersionInput, uint64) error
	ChangeDocGroup(productdomain.DocGroupInput) error
	SaveDocument(productdomain.DocumentInput, uint64) (productcontract.WebsiteDocDocument, error)
	DeleteDocument(productdomain.DocumentDeleteInput, uint64) error
	DocumentHistory(id, version string, historyID uint64) (any, error)
	SaveStorage(storage filecontract.WebsiteStorage) error
}

type websiteContentService struct {
	configs settingsservice.Store
	docs    productrepository.ProductDocsRepository
	files   filecontract.Settings
}

func NewWebsiteContentService(configs settingsservice.Store, docs productrepository.ProductDocsRepository, files filecontract.Settings) WebsiteContentService {
	return &websiteContentService{configs: configs, docs: docs, files: files}
}

func (s *websiteContentService) EnsureDefaults() error {
	if s.files != nil {
		if err := s.files.EnsureDefaults(); err != nil {
			return err
		}
	}
	// 文档只由迁移初始化。缺少新表时拒绝启动，不能回退到 JSON 或生成覆盖性默认数据。
	if _, err := s.docs.Site(false); err != nil {
		return fmt.Errorf("产品文档迁移未完成，请先运行 cmd/migrate（007_product_documents.sql）: %w", err)
	}
	defaults := []struct {
		key         string
		label       string
		description string
		value       any
		sort        int
		public      bool
	}{
		{websiteContentKey, "官网内容", "Deploy 产品官网的品牌、首页、产品与关于页内容", DefaultWebsiteContent(), 1, true},
		{websiteDownloadsKey, "下载中心", "版本、安装包、安装方式和环境要求", DefaultWebsiteDownloads(), 2, false},
	}

	for _, item := range defaults {
		existing, err := s.configs.GetConfig(item.key)
		if err != nil {
			return fmt.Errorf("读取配置 %s 失败: %w", item.key, err)
		}
		if existing != nil {
			// Storage credentials must never be exposed by generic public-config readers.
			// This also repairs rows created by earlier refactor revisions.
			if (item.key == websiteStorageKey || item.key == websiteDownloadsKey) && existing.IsPublic {
				existing.IsPublic = false
				if err := s.configs.UpdateConfig(existing); err != nil {
					return fmt.Errorf("收紧存储配置可见性失败: %w", err)
				}
			}
			continue
		}
		value, err := json.Marshal(item.value)
		if err != nil {
			return fmt.Errorf("序列化默认配置 %s 失败: %w", item.key, err)
		}
		if err := s.configs.CreateConfig(&settingsmodel.SystemConfig{
			Key: item.key, Value: string(value), Category: settingsmodel.ConfigCategoryContent,
			ConfigType: settingsmodel.ConfigTypeJSON, Label: item.label, Description: item.description,
			IsPublic: item.public, IsEditable: true, IsRequired: true, SortOrder: item.sort,
		}); err != nil {
			return fmt.Errorf("创建默认配置 %s 失败: %w", item.key, err)
		}
	}
	return nil
}

func (s *websiteContentService) GetContent() (productcontract.WebsiteContent, error) {
	value := DefaultWebsiteContent()
	return value, s.load(websiteContentKey, &value)
}

func (s *websiteContentService) GetDownloads() (productcontract.WebsiteDownloads, error) {
	value := DefaultWebsiteDownloads()
	if err := s.load(websiteDownloadsKey, &value); err != nil {
		return value, err
	}
	normalizeWebsiteDownloads(&value)
	return value, nil
}

func (s *websiteContentService) GetPublicDownloads() (productcontract.WebsiteDownloads, error) {
	value, err := s.GetDownloads()
	if err != nil {
		return value, err
	}
	// 草稿和下架包只在管理接口返回，公开接口不泄漏对应下载地址。
	visible := make([]productcontract.WebsiteDownloadPackage, 0, len(value.Packages))
	for _, item := range value.Packages {
		if item.Status == "published" && validDownloadURL(item.URL) {
			// 公开清单只给会员下载入口，实际存储地址在认证后解析。
			item.URL = "/v1/file/download?id=" + url.QueryEscape(item.ID)
			visible = append(visible, item)
		}
	}
	value.Packages = visible
	return value, nil
}

func (s *websiteContentService) SaveContent(content productcontract.WebsiteContent) error {
	if err := validateWebsiteContent(content); err != nil {
		return errorx.Invalid(err.Error())
	}
	return s.save(websiteContentKey, content)
}

func (s *websiteContentService) SaveDownloads(downloads productcontract.WebsiteDownloads) error {
	// 全量校验字段，文件存在性只校验本次新增发布、重新发布和地址变更。
	normalizeWebsiteDownloads(&downloads)
	if err := validateWebsiteDownloads(downloads); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidWebsiteDownloads, err)
	}
	snapshot, err := s.configs.GetConfig(websiteDownloadsKey)
	if err != nil {
		return err
	}
	if snapshot == nil {
		if err := s.EnsureDefaults(); err != nil {
			return err
		}
		snapshot, err = s.configs.GetConfig(websiteDownloadsKey)
		if err != nil {
			return err
		}
		if snapshot == nil {
			return errors.New("下载配置初始化后仍不存在")
		}
	}
	previous := DefaultWebsiteDownloads()
	if strings.TrimSpace(snapshot.Value) != "" {
		if err := json.Unmarshal([]byte(snapshot.Value), &previous); err != nil {
			return fmt.Errorf("读取已有下载配置失败: %w", err)
		}
	}
	normalizeWebsiteDownloads(&previous)
	// 文件系统 I/O 在数据库锁外完成；提交时锁定同一配置并重验快照，
	// 防止校验期间清单变化后仍按旧状态跳过发布检查。
	if err := s.validateLocalReleases(previous, downloads); err != nil {
		return err
	}
	encoded, err := json.Marshal(downloads)
	if err != nil {
		return fmt.Errorf("序列化下载内容失败: %w", err)
	}
	return s.configs.UpdateJSONConfigIfUnchanged(websiteDownloadsKey, snapshot.Value, string(encoded))
}

func (s *websiteContentService) load(key string, target any) error {
	config, err := s.configs.GetConfig(key)
	if err != nil {
		return err
	}
	if config == nil || strings.TrimSpace(config.Value) == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(config.Value), target); err != nil {
		return fmt.Errorf("配置 %s 不是有效 JSON: %w", key, err)
	}
	return nil
}

func (s *websiteContentService) save(key string, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化官网内容失败: %w", err)
	}
	config, err := s.configs.GetConfig(key)
	if err != nil {
		return err
	}
	if config == nil {
		if err := s.EnsureDefaults(); err != nil {
			return err
		}
	}
	return s.configs.BatchUpdateConfigs([]settingsservice.BatchUpdateConfig{{
		Key: key, Value: string(encoded), ConfigType: string(settingsmodel.ConfigTypeJSON),
	}})
}

func validateWebsiteContent(content productcontract.WebsiteContent) error {
	if strings.TrimSpace(content.Brand.Name) == "" || strings.TrimSpace(content.Home.Title) == "" {
		return fmt.Errorf("品牌名称和首页标题不能为空")
	}
	if content.Brand.ContactEmail != "" {
		if _, err := mail.ParseAddress(content.Brand.ContactEmail); err != nil {
			return fmt.Errorf("联系邮箱格式不正确")
		}
	}
	if len(content.Capabilities) == 0 {
		return fmt.Errorf("至少保留一个产品能力")
	}
	return nil
}

func validateWebsiteDownloads(downloads productcontract.WebsiteDownloads) error {
	groups := make(map[string]struct{}, len(downloads.Groups))
	for _, group := range downloads.Groups {
		if group.ID == "" || group.ID == "ungrouped" || len(group.ID) > 200 || group.Name == "" || len([]rune(group.Name)) > 80 {
			return fmt.Errorf("下载分组需要有效标识和不超过 80 字的名称")
		}
		if _, exists := groups[group.ID]; exists {
			return fmt.Errorf("下载分组标识重复")
		}
		groups[group.ID] = struct{}{}
	}
	ids := make(map[string]struct{}, len(downloads.Packages))
	for _, item := range downloads.Packages {
		if item.GroupID != "" {
			if _, exists := groups[item.GroupID]; !exists {
				return fmt.Errorf("文件 %s 的下载分组不存在，请先调整文件所属分组", item.ID)
			}
		}
		if len([]rune(item.Edition)) > 80 || item.SortOrder < 0 || item.SortOrder > 99999 {
			return fmt.Errorf("版本类型不能超过 80 字，排序应在 0 到 99999 之间")
		}
		if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Platform) == "" {
			return fmt.Errorf("安装包 ID 和平台名称不能为空")
		}
		if _, exists := ids[item.ID]; exists {
			return fmt.Errorf("安装包 ID %s 重复", item.ID)
		}
		ids[item.ID] = struct{}{}
		if item.Status != "draft" && item.Status != "published" && item.Status != "archived" {
			return fmt.Errorf("安装包 %s 状态无效", item.ID)
		}
		if item.Size < 0 {
			return fmt.Errorf("安装包大小不能为负数")
		}
		if !ValidDownloadName(item.FileName) {
			return fmt.Errorf("下载文件名不能包含路径、特殊符号或超过 255 字节")
		}
		if item.URL != "" && !validDownloadURL(item.URL) {
			return fmt.Errorf("下载地址仅支持 HTTP、HTTPS 或站内绝对路径")
		}
		if item.Status == "published" && (item.URL == "" || strings.TrimSpace(item.Version) == "") {
			return fmt.Errorf("发布安装包前请填写版本号和下载地址")
		}
		if item.Checksum != "" {
			if decoded, err := hex.DecodeString(item.Checksum); err != nil || len(decoded) != 32 {
				return fmt.Errorf("SHA256 必须为 64 位十六进制字符")
			}
		}
	}
	return nil
}

func validDownloadURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || raw == "" || strings.ContainsAny(raw, "\\\r\n\t") {
		return false
	}
	if parsed.Scheme == "" {
		return strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") && parsed.Host == ""
	}
	return (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.Host != "" && parsed.User == nil
}

func normalizeWebsiteDownloads(value *productcontract.WebsiteDownloads) {
	value.Title = strings.TrimSpace(value.Title)
	if value.Title == "" {
		value.Title = "下载中心"
	}
	if value.Groups == nil {
		value.Groups = []productcontract.WebsiteDownloadGroup{}
	}
	for i := range value.Groups {
		value.Groups[i].ID = strings.TrimSpace(value.Groups[i].ID)
		value.Groups[i].Name = strings.TrimSpace(value.Groups[i].Name)
		value.Groups[i].Description = strings.TrimSpace(value.Groups[i].Description)
	}
	for i := range value.Packages {
		item := &value.Packages[i]
		item.GroupID = strings.TrimSpace(item.GroupID)
		item.Edition = strings.TrimSpace(item.Edition)
		item.ID = strings.TrimSpace(item.ID)
		item.Platform = strings.TrimSpace(item.Platform)
		item.URL = normalizeReleaseURL(item.URL)
		item.Checksum = strings.TrimSpace(item.Checksum)
		if item.Product == "" {
			item.Product = "Deploy"
		}
		if item.Version == "" {
			item.Version = value.Version
		}
		if item.ReleaseDate == "" {
			item.ReleaseDate = value.ReleaseDate
		}
	}
}

func DefaultWebsiteContent() productcontract.WebsiteContent {
	return productcontract.WebsiteContent{
		Brand: productcontract.WebsiteBrand{
			Name: "Deploy", FullName: "Deploy 企业一体化 DevOps 运维平台",
			Description: "统一资产、自动化发布、远程运维与审计，帮助技术团队更快、更稳、更安全地交付应用。",
			Copyright:   "Deploy. All rights reserved.",
		},
		Navigation: []productcontract.WebsiteNavigation{{Label: "首页", Path: "/"}, {Label: "产品", Path: "/product"}, {Label: "下载", Path: "/download"}, {Label: "文档", Path: "/docs"}, {Label: "关于", Path: "/about"}},
		Home: productcontract.WebsiteHomeContent{
			Eyebrow: "为自建基础设施而设计", Title: "让部署、运维与", Highlight: "安全管理", Description: "统一资产、自动化发布、远程运维与审计，帮助技术团队更快、更稳、更安全地交付应用。",
			PrimaryAction: "获取 Deploy", SecondaryAction: "了解产品能力", Meta: []string{"Server + K8s", "Playbook + Pipeline", "RBAC + Audit"},
			CapabilityTitle: "一个控制面，接住交付链路上的每个关键动作", CapabilityIntro: "Deploy 将基础设施、发布流程和安全治理放进同一上下文，让团队看到一条完整链路，而不是更多孤立工具。",
			FlowTitle: "把复杂发布，变成清晰、可控的交付路径", FlowIntro: "代码、制品、环境、审批与操作记录天然关联，每一步都能回到同一条发布记录。",
			AudienceTitle: "面向真实团队与复杂环境", AudienceIntro: "适用于拥有自建服务器或 Kubernetes 环境的软件企业、政企 IT 与交付团队。",
			BlogTitle: "来自 Deploy 团队的实践与洞察", CTA: "从下一次发布开始，把复杂留给 Deploy。",
		},
		Capabilities: []productcontract.WebsiteCapability{
			{Key: "assets", Index: "01", Title: "统一资产管理", Description: "集中纳管服务器、Kubernetes 集群、主机账号与运行环境，让资产状态和责任边界始终清晰。", Tags: []string{"Server", "Kubernetes", "环境分组"}},
			{Key: "pipeline", Index: "02", Title: "可视化 CI/CD", Description: "编排代码拉取、构建、制品、审批和发布阶段，用一条流水线贯通多环境交付。", Tags: []string{"Pipeline", "多环境", "审批"}},
			{Key: "playbook", Index: "03", Title: "Playbook 自动化", Description: "将重复运维动作沉淀为可复用、可参数化、可版本管理的标准作业。", Tags: []string{"批量执行", "版本管理", "执行记录"}},
			{Key: "artifact", Index: "04", Title: "制品与发布", Description: "关联代码版本、构建制品与目标环境，支持审批、灰度、健康检查与发布回溯。", Tags: []string{"制品库", "灰度发布", "回滚"}},
			{Key: "terminal", Index: "05", Title: "安全远程运维", Description: "通过统一受控入口访问目标资产，将身份、会话和命令操作纳入平台治理。", Tags: []string{"受控访问", "会话记录", "文件传输"}},
			{Key: "shield", Index: "06", Title: "权限与审计", Description: "使用细粒度 RBAC、审批策略与完整操作日志，守住生产环境的每一次变更。", Tags: []string{"RBAC", "操作审计", "安全策略"}},
		},
		DeliverySteps: []productcontract.WebsiteDeliveryStep{
			{Index: "01", Key: "CODE", Title: "接入代码", Description: "连接代码仓库，配置分支与触发策略。"}, {Index: "02", Key: "BUILD", Title: "构建制品", Description: "标准化构建过程，统一归档版本制品。"},
			{Index: "03", Key: "RELEASE", Title: "环境发布", Description: "审批后投放到服务器或 Kubernetes。"}, {Index: "04", Key: "AUDIT", Title: "回溯审计", Description: "状态、操作者和结果在同一记录中可查。"},
		},
		Audiences: []productcontract.WebsiteAudience{
			{Index: "01", Title: "软件企业", Headline: "把零散脚本变成标准交付能力", Description: "统一研发、测试与生产环境的发布方式，让小团队也能建立稳定、可复制的 DevOps 流程。"},
			{Index: "02", Title: "政企 IT", Headline: "在合规边界内提升运维效率", Description: "权限、命令执行和生产变更集中留痕，让安全要求成为流程的一部分。"},
			{Index: "03", Title: "交付团队", Headline: "用一套方法管理多客户环境", Description: "按项目和环境组织资产与模板，让实施经验可以复用，交付现场不再从头搭建。"},
		},
		Product: productcontract.WebsiteProductContent{
			Title: "为企业交付构建的一体化 DevOps 平台", Description: "从基础设施纳管到应用发布，再到权限与审计，Deploy 用统一模型连接日常运维中的关键环节。",
			CapabilityTitle: "覆盖交付与运维的六个核心能力", CapabilityIntro: "各模块共享资产、身份和审计上下文，既可以独立使用，也能组合成完整的交付流程。",
			ArchitectureTitle: "一套平台连接代码、制品与基础设施", ArchitectureIntro: "Deploy 不改变团队已有的代码仓库和运行环境，而是在它们之间建立标准化的交付控制面。",
			ArchitectureLayers: []productcontract.WebsiteArchitectureLayer{{Name: "接入层", Items: []string{"Git 仓库", "镜像仓库", "服务器", "Kubernetes"}}, {Name: "平台层", Items: []string{"资产中心", "Playbook", "CI/CD", "制品管理", "权限中心", "审计中心"}}, {Name: "交付层", Items: []string{"开发环境", "测试环境", "预发布环境", "生产环境"}}},
			SecurityTitle:      "让每一次访问都有身份、有边界、有记录", SecurityIntro: "安全不是流程外的额外环节，而是贯穿资产访问、自动化作业和应用发布的默认规则。",
			SecurityItems: []productcontract.WebsiteSecurityItem{{Title: "身份与权限", Description: "按团队、角色和资源范围分配权限，将最小权限落实到具体资产和动作。"}, {Title: "变更可追溯", Description: "发布记录、审批上下文和操作日志集中关联，问题回溯不再拼凑证据。"}, {Title: "运维有边界", Description: "远程会话通过受控入口完成，敏感命令和文件传输可以记录并复盘。"}},
		},
		About: productcontract.WebsiteAboutContent{
			Title: "让复杂运维，成为可管理的工程系统", Description: "面向拥有自建服务器或 Kubernetes 环境的软件企业与政企 IT 团队。", IntroTitle: "交付不该散落在脚本、聊天记录和个人经验里",
			Paragraphs: []string{"当服务器、集群、流水线和运维入口分别由不同工具承载，环境信息与操作责任很容易失去上下文。", "Deploy 将资产、自动化发布、远程运维、权限和审计放到同一个平台，帮助团队建立清晰、可复用、可追溯的交付方式。"},
			Principles: []productcontract.WebsitePrinciple{{Index: "01", Title: "以交付结果为中心", Description: "从代码、构建、制品到多环境发布，围绕一条可追踪的交付链路组织能力。"}, {Index: "02", Title: "自动化但不失控", Description: "将重复动作沉淀为标准流程，同时保留审批、权限、审计与人工确认边界。"}, {Index: "03", Title: "适配真实基础设施", Description: "同时面向传统服务器与 Kubernetes 环境，让团队可以按自己的节奏演进。"}},
			TeamTitle:  "为需要对生产环境负责的人而设计", TeamIntro: "Deploy 关注企业基础设施里的实际协作：谁可以访问、谁批准发布、发生了什么，以及如何快速回溯。", TeamRoles: []string{"运维工程师", "DevOps 工程师", "系统管理员", "交付实施人员", "技术负责人"},
		},
		Footer: productcontract.WebsiteFooterContent{Slogan: "让交付更快、更稳、更安全。", Statement: "用确定的流程，交付不确定的未来。"},
	}
}

func DefaultWebsiteDownloads() productcontract.WebsiteDownloads {
	return productcontract.WebsiteDownloads{
		Version: "即将发布", ReleaseDate: "待公布", SupportArchitectures: "AMD64 / ARM64", Description: "选择适合基础设施的安装方式。下载地址由企业发布流程统一配置，正式发布前不会提供无效安装包。",
		Packages:       []productcontract.WebsiteDownloadPackage{{ID: "linux-amd64", Platform: "Linux", Arch: "x86_64 / AMD64", Format: "二进制安装包", Icon: "LINUX"}, {ID: "linux-arm64", Platform: "Linux", Arch: "ARM64", Format: "二进制安装包", Icon: "ARM"}, {ID: "offline-amd64", Platform: "离线部署包", Arch: "x86_64 / AMD64", Format: "镜像与依赖集合", Icon: "OFFLINE"}},
		InstallOptions: []productcontract.WebsiteInstallOption{{ID: "docker", Title: "Docker Compose", Badge: "快速体验", Description: "适合单机体验、功能评估和小规模团队内部使用。", Command: "docker compose up -d"}, {ID: "linux", Title: "Linux 安装包", Badge: "推荐", Description: "适合自建服务器环境，按节点规划独立部署与持久化目录。", Command: "chmod +x deploy-installer && sudo ./deploy-installer"}, {ID: "kubernetes", Title: "Kubernetes / Helm", Badge: "集群部署", Description: "适合已有 Kubernetes 的团队，便于统一资源编排和弹性扩展。", Command: "helm upgrade --install deploy ./deploy-chart -n deploy --create-namespace"}},
		Requirements:   []productcontract.WebsiteRequirement{{Label: "操作系统", Value: "主流 64 位 Linux 发行版"}, {Label: "容器环境", Value: "Docker / containerd"}, {Label: "Kubernetes", Value: "生产环境建议使用受支持版本"}, {Label: "浏览器", Value: "Chrome / Edge / Firefox 最新两个大版本"}, {Label: "网络", Value: "允许平台访问被纳管资产与镜像仓库"}, {Label: "安全", Value: "生产环境必须配置 HTTPS 与独立密钥"}},
		SupportTitle:   "需要私有化部署与实施支持？", SupportDescription: "企业版面向正式生产环境提供部署规划、升级策略与技术支持。",
	}
}

// Compatibility methods preserve the current HTTP contract; file owns the data.
func (s *websiteContentService) GetStorage() (filecontract.WebsiteStorage, error) {
	return s.files.GetStorage()
}
func (s *websiteContentService) SaveStorage(value filecontract.WebsiteStorage) error {
	return s.files.SaveStorage(value)
}
