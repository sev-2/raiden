package builder

import "strings"

// StorageBucketClause returns the canonical bucket scope clause for Supabase storage policies.
func StorageBucketClause(bucketName string) Clause {
	bucketName = strings.TrimSpace(bucketName)
	if bucketName == "" {
		return ""
	}
	return EqString("bucket_id", bucketName)
}

// StorageUsingClause ensures the storage policy USING clause is scoped to the bucket.
// Additional constraints are AND-ed to the required bucket filter, preserving pre-existing filters.
func StorageUsingClause(bucketName string, clause Clause) Clause {
	return combineStorageClause(bucketName, clause)
}

// StorageCheckClause ensures the storage policy CHECK clause is scoped to the bucket.
// Additional constraints are AND-ed to the required bucket filter, preserving pre-existing filters.
func StorageCheckClause(bucketName string, clause Clause) Clause {
	return combineStorageClause(bucketName, clause)
}

// StripStorageBucketFilter removes the default bucket filter from a raw clause so that
// generator routines can rebuild it using StorageUsingClause or StorageCheckClause.
func StripStorageBucketFilter(sql string, bucketName string) string {
	normalized := strings.TrimSpace(NormalizeClauseSQL(sql))
	if normalized == "" {
		return ""
	}

	bucketClause := strings.TrimSpace(NormalizeClauseSQL(StorageBucketClause(bucketName).String()))
	if bucketClause == "" {
		return normalized
	}

	normalized = strings.ReplaceAll(normalized, "("+bucketClause+")", bucketClause)

	if strings.EqualFold(normalized, bucketClause) {
		return ""
	}

	if parts := splitByLogical(normalized, "AND"); parts != nil {
		filtered := filterStorageClause(parts, bucketClause)
		if len(filtered) == 0 {
			return ""
		}
		if len(filtered) < len(parts) {
			return strings.Join(filtered, " AND ")
		}
	}

	if parts := splitByLogical(normalized, "OR"); parts != nil {
		filtered := filterStorageClause(parts, bucketClause)
		if len(filtered) == 0 {
			return ""
		}
		if len(filtered) < len(parts) {
			return strings.Join(filtered, " OR ")
		}
	}

	if strings.EqualFold(normalized, bucketClause) {
		return ""
	}

	return normalized
}

func combineStorageClause(bucketName string, clause Clause) Clause {
	bucketClause := StorageBucketClause(bucketName)
	if bucketClause.IsEmpty() {
		return clause
	}

	normalizedClause := strings.ToLower(strings.TrimSpace(NormalizeClauseSQL(clause.String())))
	stripped := strings.ToLower(strings.TrimSpace(StripStorageBucketFilter(clause.String(), bucketName)))
	if normalizedClause != "" && stripped != normalizedClause {
		return clause
	}

	if clause.IsEmpty() {
		return bucketClause
	}

	return And(bucketClause, clause)
}

func filterStorageClause(parts []string, bucketClause string) []string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		canonical := trimOuterParens(trimmed)
		normalizedPart := strings.TrimSpace(NormalizeClauseSQL(canonical))
		if strings.EqualFold(normalizedPart, bucketClause) {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func trimOuterParens(s string) string {
	result := strings.TrimSpace(s)
	for len(result) >= 2 && result[0] == '(' && result[len(result)-1] == ')' {
		inner := strings.TrimSpace(result[1 : len(result)-1])
		if inner == "" {
			return ""
		}
		result = inner
	}
	return result
}
