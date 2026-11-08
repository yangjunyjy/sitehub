// Older servers may return duplicate role relations and nickName instead of nickname.
export function normalizeUserRow(row) {
  const seen = new Set()
  const roles = (Array.isArray(row.roles) ? row.roles : []).filter((role) => {
    if (!role || role.id == null || seen.has(String(role.id))) return false
    seen.add(String(role.id))
    return true
  })
  const { password, ...user } = row
  return {
    ...user,
    nickname: row.nickname ?? row.nickName ?? '',
    roles,
    roleIds: roles.map((role) => role.id),
  }
}
