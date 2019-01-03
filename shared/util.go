package shared

import "strings"

func GetKey(key string, args []string) (string, bool) {
	var (
		v     string
		exist bool
	)

	for _, value := range args {
		s := strings.Split(value, "=")
		if len(s) != 2 {
			continue
		}
		if s[0] == "-"+key {
			exist = true
			v = s[1]
		}
	}

	return v, exist
}

func NormalizeKey(key string) string {
	return strings.Replace(key, " ", "_", -1)
}
