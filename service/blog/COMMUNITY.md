# 博客社区内容

部署已有博客库时运行项目的 `bootstrap --service blog` 迁移。版本 2 会创建 `friend_links`、`guest_messages` 和 `author_profiles`，保留已有文章和分类。

版本 4 将现有留言保留在留言板，新文章评论通过 `article_id` 关联到文章。升级已有云部署时，使用新版发布包运行 `bash init.sh database`，同时补齐后台菜单。

友链申请、留言和文章评论提交后进入待审核列表。在「博客系统 → 友链管理」选择相应页签，审核通过后会出现在对应公开页面。「已公开」页签支持撤下内容。其他管理角色需分配该菜单以及 `blog:community:list` 和 `blog:community:review` 权限。

关于页读取 `author_profiles` 中 `id = 1` 的公开资料。`skills` 是字符串数组，`socials` 是含 `label` 和 `url` 的对象数组。首次设置可在博客库执行以下示例并替换内容。

```sql
INSERT INTO author_profiles (id, name, title, bio, content, skills, socials)
VALUES (1, '博主', '关于我', '这里记录我的学习与生活。', '', JSON_ARRAY('Go'), JSON_ARRAY());
```
