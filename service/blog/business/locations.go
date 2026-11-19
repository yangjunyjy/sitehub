package business

import filecontract "sitehub/service/blog/contract"

func normalizeReleaseURL(raw string) string { return filecontract.NormalizeReleaseURL(raw) }
