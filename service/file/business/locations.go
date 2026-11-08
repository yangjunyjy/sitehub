package services

import filecontract "sitehub/service/file/contract"

func normalizeReleaseURL(raw string) string { return filecontract.NormalizeReleaseURL(raw) }
