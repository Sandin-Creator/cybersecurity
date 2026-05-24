package report

import (
	"fmt"
	"strings"
)

// PrintBanner returns an ASCII art banner for the tool
func GetBanner() string {
	return `
    ____  _       _ _        _    ____       _        _   _           
   |  _ \(_) __ _(_) |_ __ _| |  |  _ \  ___| |_ ___| |_(_)_   _____ 
   | | | | |/ _` + "`" + ` | | __/ _` + "`" + ` | |  | | | |/ _ \ __/ _ \ __| | \ \ / / _ \
   | |_| | | (_| | | || (_| | |  | |_| |  __/ ||  __/ |_| | |\ V /  __/
   |____/|_|\__, |_|\__\__,_|_|  |____/ \___|\__\___|\__|_|_| \_/ \___|
            |___/                                                      
`
}

// FormatSection creates a boxed section for the report
func FormatSection(title, content string) string {
	var sb strings.Builder
	line := strings.Repeat("=", 60)
	
	sb.WriteString("\n" + line + "\n")
	sb.WriteString(fmt.Sprintf("| %-56s |\n", strings.ToUpper(title)))
	sb.WriteString(line + "\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(line + "\n")
	return sb.String()
}
