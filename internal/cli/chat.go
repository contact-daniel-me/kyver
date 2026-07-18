package cli

import (
	"fmt"

	"github.com/contact-daniel-me/kyver/internal/growiota"
	"github.com/contact-daniel-me/kyver/internal/growiota/chat"
)

func runChat(args []string) error {
	if !growiota.IsActivated() {
		fmt.Println("=================================================================")
		fmt.Println(" AI capabilities are available through the Growiota Premium module.")
		fmt.Println(" To activate, set GROWIOTA_PREMIUM=1.")
		fmt.Println(" All core Kyver local intelligence features (Search, Context, ")
		fmt.Println(" Retrieval, etc.) remain fully functional offline.")
		fmt.Println("=================================================================")
		return nil
	}

	return chat.RunChat(args)
}
