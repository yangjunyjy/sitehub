package services

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"slices"
	"strings"

	errorx "sitehub/common/errorx"
	settingsservice "sitehub/common/settings/contract"
	settingsmodel "sitehub/common/settings/model"
	productcontract "sitehub/service/product/contract"
)

const websiteContentKey = "website.content"

type WebsiteContentService interface {
	EnsureDefaults() error
	GetContent() (productcontract.WebsiteContent, error)
	SaveContent(productcontract.WebsiteContent) error
}

type websiteContentService struct {
	configs settingsservice.Store
}

func NewWebsiteContentService(store settingsservice.Store) WebsiteContentService {
	return &websiteContentService{configs: store}
}

func (s *websiteContentService) EnsureDefaults() error {
	row, err := s.configs.GetConfig(websiteContentKey)
	if err != nil || row != nil {
		return err
	}
	raw, err := json.Marshal(DefaultWebsiteContent())
	if err != nil {
		return err
	}
	return s.configs.CreateConfig(&settingsmodel.SystemConfig{
		Key: websiteContentKey, Value: string(raw), Category: settingsmodel.ConfigCategoryContent,
		ConfigType: settingsmodel.ConfigTypeJSON, Label: "官网内容", Description: "品牌、首页和产品介绍",
		IsPublic: true, IsEditable: true, IsRequired: true, SortOrder: 1,
	})
}

func (s *websiteContentService) GetContent() (productcontract.WebsiteContent, error) {
	defaults := DefaultWebsiteContent()
	content := defaults
	if err := s.load(websiteContentKey, &content); err != nil {
		return content, err
	}
	// Existing installations persist the original seed. Adapt matching fields for display
	// without writing settings during a read or replacing edited copy from the same site.
	content.Home.ParticleShape = publishedParticleShape(content.Home.ParticleShape)
	for _, field := range []struct {
		value        *string
		old, current string
	}{
		{&content.Home.Eyebrow, "为自建基础设施而设计", defaults.Home.Eyebrow},
		{&content.Home.Title, "让部署、运维与", defaults.Home.Title},
		{&content.Home.Highlight, "安全管理", defaults.Home.Highlight},
		{&content.Home.Description, "统一资产、自动化发布、远程运维与审计，帮助技术团队更快、更稳、更安全地交付应用。", defaults.Home.Description},
		{&content.Home.PrimaryAction, "获取 Deploy", defaults.Home.PrimaryAction},
		{&content.Home.SecondaryAction, "了解产品能力", defaults.Home.SecondaryAction},
		{&content.Home.CapabilityTitle, "一个控制面，接住交付链路上的每个关键动作", defaults.Home.CapabilityTitle},
		{&content.Home.CapabilityIntro, "Deploy 将基础设施、发布流程和安全治理放进同一上下文，让团队看到一条完整链路，而不是更多孤立工具。", defaults.Home.CapabilityIntro},
		{&content.Home.FlowTitle, "把复杂发布，变成清晰、可控的交付路径", defaults.Home.FlowTitle},
		{&content.Home.FlowIntro, "代码、制品、环境、审批与操作记录天然关联，每一步都能回到同一条发布记录。", defaults.Home.FlowIntro},
		{&content.Home.AudienceTitle, "面向真实团队与复杂环境", defaults.Home.AudienceTitle},
		{&content.Home.AudienceIntro, "适用于拥有自建服务器或 Kubernetes 环境的软件企业、政企 IT 与交付团队。", defaults.Home.AudienceIntro},
		{&content.Home.BlogTitle, "来自 Deploy 团队的实践与洞察", defaults.Home.BlogTitle},
		{&content.Home.CTA, "从下一次发布开始，把复杂留给 Deploy。", defaults.Home.CTA},
		{&content.Footer.Slogan, "让交付更快、更稳、更安全。", defaults.Footer.Slogan},
		{&content.Footer.Statement, "用确定的流程，交付不确定的未来。", defaults.Footer.Statement},
	} {
		if *field.value == field.old {
			*field.value = field.current
		}
	}
	if slices.Equal(content.Home.Meta, []string{"Server + K8s", "Playbook + Pipeline", "RBAC + Audit"}) {
		content.Home.Meta = []string{}
	}
	return content, nil
}

func (s *websiteContentService) SaveContent(content productcontract.WebsiteContent) error {
	content.Home.ParticleShape = publishedParticleShape(content.Home.ParticleShape)
	if err := validateWebsiteContent(content); err != nil {
		return errorx.Invalid(err.Error())
	}
	return s.save(websiteContentKey, content)
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

// Preserve the selected preset and both motion preferences on existing sites.
func publishedParticleShape(shape string) string {
	switch shape {
	case "ring":
		return "artifact"
	case "wave":
		return "cluster"
	case "helix":
		return "terminal"
	default:
		return shape
	}
}

func validateWebsiteContent(content productcontract.WebsiteContent) error {
	switch content.Home.ParticleShape {
	case "artifact", "cluster", "terminal":
	default:
		return fmt.Errorf("请选择有效的首页粒子形态")
	}
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

func DefaultWebsiteContent() productcontract.WebsiteContent {
	return productcontract.WebsiteContent{
		Brand: productcontract.WebsiteBrand{
			Name: "Deploy", FullName: "Deploy 企业一体化 DevOps 运维平台",
			Description: "统一资产、自动化发布、远程运维与审计，帮助技术团队更快、更稳、更安全地交付应用。",
			Copyright:   "Deploy. All rights reserved.",
		},
		Navigation: []productcontract.WebsiteNavigation{{Label: "首页", Path: "/"}, {Label: "产品", Path: "/product"}, {Label: "下载", Path: "/download"}, {Label: "文档", Path: "/docs"}, {Label: "关于", Path: "/about"}},
		Home: productcontract.WebsiteHomeContent{
			ParticleShape: "artifact", ParticleAutoplay: true, ParticleInteractive: true,
			Title: "部署与运维", PrimaryAction: "下载", SecondaryAction: "文档", Meta: []string{},
			CapabilityTitle: "产品功能", FlowTitle: "交付流程", AudienceTitle: "适用团队",
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
	}
}
