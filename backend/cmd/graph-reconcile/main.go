// graph-reconcile compares legacy module graph snapshots with normalized graph
// rows. It never runs migrations and reconciliation queries are read-only.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"backend/config"
	"backend/db"
	domainModule "backend/internal/domain/module"
	usecaseModule "backend/internal/usecase/module"
)

type output struct {
	Reports []domainModule.GraphReconciliationReport `json:"reports"`
}

func main() {
	moduleID := flag.String("module-id", "", "module ID to reconcile")
	projectID := flag.String("project-id", "", "project ID whose modules will be reconciled")
	flag.Parse()
	if (*moduleID == "" && *projectID == "") || (*moduleID != "" && *projectID != "") {
		fmt.Fprintln(os.Stderr, "provide exactly one of --module-id or --project-id")
		os.Exit(2)
	}

	config.InitConfig()
	if err := db.InitReadOnly(); err != nil {
		log.Fatalf("graph reconciliation connection failed: %v", err)
	}
	defer db.DB.Close()

	ctx := context.Background()
	var reports []domainModule.GraphReconciliationReport
	var err error
	if *moduleID != "" {
		var report domainModule.GraphReconciliationReport
		report, err = usecaseModule.ExecuteReconcileGraph(ctx, db.DB, *moduleID)
		reports = []domainModule.GraphReconciliationReport{report}
	} else {
		reports, err = usecaseModule.ExecuteReconcileProject(ctx, db.DB, *projectID)
	}
	if err != nil {
		log.Fatalf("graph reconciliation failed: %v", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output{Reports: reports}); err != nil {
		log.Fatalf("write graph reconciliation report: %v", err)
	}
}
