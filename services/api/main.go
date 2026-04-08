package main

import (
	"os"

	"github.com/decisionbox-io/decisionbox/services/api/apiserver"
	"github.com/decisionbox-io/decisionbox/services/api/internal/backfill"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "backfill-embeddings" {
		backfill.RunBackfillEmbeddings(os.Args[2:])
		return
	}

	apiserver.Run()
}
