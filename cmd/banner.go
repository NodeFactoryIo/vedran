package cmd

import "fmt"

const banner = "\n _    ____________  ____  ___    _   __\n| |  / / ____/ __ \\/ __ \\/   |  / | / /\n| | / / __/ / / / / /_/ / /| | /  |/ / \n| |/ / /___/ /_/ / _, _/ ___ |/ /|  /  \n|___/_____/_____/_/ |_/_/  |_/_/ |_/   \n                                       \n"

func DisplayBanner() {
	fmt.Print(banner)
}