package raiden

import (
	"regexp"
	"strings"
)

type (
	Acl struct {
		Roles []string
		Check *string
		Using string
	}

	AclTag struct {
		Read  Acl
		Write Acl
	}
)

func UnmarshalAclTag(tag string) AclTag {
	var aclTag AclTag

	aclTagMap := make(map[string]string)

	// Regular expression to match key-value pairs
	re := regexp.MustCompile(`(\w+):"([^"]*)"`)

	// Find all matches
	matches := re.FindAllStringSubmatch(tag, -1)

	// Loop through matches and add to result map
	for _, match := range matches {
		if len(match) == 3 {
			key := match[1]
			value := match[2]
			aclTagMap[key] = value
		}
	}

	if readTag, exist := aclTagMap["read"]; exist && len(readTag) > 0 {
		aclTag.Read.Roles = strings.Split(readTag, ",")
	}

	if writeTag, exist := aclTagMap["write"]; exist && len(writeTag) > 0 {
		aclTag.Write.Roles = strings.Split(writeTag, ",")
	}

	if readTagUsing, exist := aclTagMap["readUsing"]; exist && len(readTagUsing) > 0 {
		aclTag.Read.Using = readTagUsing
	}

	if writeTagCheck, exist := aclTagMap["writeCheck"]; exist && len(writeTagCheck) > 0 {
		aclTag.Write.Check = &writeTagCheck
	}

	if writeTagUsing, exist := aclTagMap["writeUsing"]; exist && len(writeTagUsing) > 0 {
		aclTag.Write.Using = writeTagUsing
	}

	return aclTag
}
