package devdesk

import (
	"fmt"
	"io"
)

const devDeskBanner = ` ____             ____             _    
|  _ \  _____   _|  _ \  ___  ___| | __
| | | |/ _ \ \ / / | | |/ _ \/ __| |/ /
| |_| |  __/\ V /| |_| |  __/\__ \   < 
|____/ \___| \_/ |____/ \___||___/_|\_\
`

func printBanner(stdout io.Writer) {
	fmt.Fprint(stdout, devDeskBanner)
	fmt.Fprintln(stdout)
}
