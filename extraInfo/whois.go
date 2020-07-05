package extrainfo

import (
	"strings"
	"fmt"
	"github.com/likexian/whois-go"
)

func formatRaw(raw string, lookFor string) string {
	format := strings.Split(raw, "\n")
	var result string
	for _, value := range format {
		if !strings.HasPrefix(value, lookFor) {
			continue
		}
		subs := strings.Split(value, ":")
		var index int
		found := false
		for pos, char := range subs[1] {
			if char != ' ' {
				index = pos
				found = true
				break
			}
		}
		if found == true {
			result = subs[1][index:]
			break
		}
	}
	return result
}

func GetWhoIsData(ip string) (string, string) {
	raw, err := whois.Whois(ip)
	var name string
	var country string
	if err == nil {
		name = formatRaw(raw, "OrgName")
		country = formatRaw(raw, "Country")

	}else{
		fmt.Println("error in whois: ", err)
	}
	
	return name, country
}
