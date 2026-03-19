package console

import (
	"database/sql"
	"strconv"

	"github.com/Back-Seat-Boy/kumpul-be/internal/config"
	"github.com/kumparan/go-utils"
	_ "github.com/lib/pq" // import postgres driver
	migrate "github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database",
	Long:  `This subcommand used to migrate database`,
	Run:   processMigration,
}

func init() {
	migrateCmd.PersistentFlags().Int("step", 0, "maximum migration steps")
	migrateCmd.PersistentFlags().String("direction", "up", "migration direction")
	RootCmd.AddCommand(migrateCmd)
}

func processMigration(cmd *cobra.Command, _ []string) {
	direction := cmd.Flag("direction").Value.String()
	stepStr := cmd.Flag("step").Value.String()
	step, err := strconv.Atoi(stepStr)
	if err != nil {
		log.WithField("stepStr", stepStr).Fatal("Failed to parse step to int: ", err)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "./db/migrations",
	}

	migrate.SetTable("schema_migrations")
	db, err := sql.Open("postgres", config.DBDSN())
	if err != nil {
		log.WithField("DatabaseDSN", config.DBDSN()).Fatal("Failed to connect database: ", err)
	}

	var n int
	if direction == "down" {
		n, err = migrate.ExecMax(db, "postgres", migrations, migrate.Down, step)
	} else {
		n, err = migrate.ExecMax(db, "postgres", migrations, migrate.Up, step)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"db":         db,
			"migrations": utils.Dump(migrations),
			"direction":  direction}).
			Fatal("Failed to migrate database: ", err)
	}

	log.Infof("Applied %d migrations!\n", n)

}
