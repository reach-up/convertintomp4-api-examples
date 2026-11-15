// Usage: go run . <example>   (01 to 06)
package main

import (
	"fmt"
	"os"
)

func main() {
	examples := map[string]func(){
		"01": tasksAndJobs,
		"02": checkStatusWithPolling,
		"03": realtimeWebsocketUpdates,
		"04": uploadingFiles,
		"05": handlingResultsAndErrors,
		"06": complexJob,
	}
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Set CIM4_API_KEY to an API key from https://convertintomp4.com/account/api/keys")
		os.Exit(1)
	}
	if len(os.Args) < 2 || examples[os.Args[1]] == nil {
		fmt.Fprintln(os.Stderr, "usage: go run . 01|02|03|04|05|06")
		os.Exit(2)
	}
	examples[os.Args[1]]()
}
