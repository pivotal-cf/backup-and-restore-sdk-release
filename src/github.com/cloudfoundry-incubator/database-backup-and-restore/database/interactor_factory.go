package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type BackuperFactory struct {
	utilitiesConfig               config.UtilitiesConfig
	postgresServerVersionDetector ServerVersionDetector
}

func NewInteractorFactory(
	utilitiesConfig config.UtilitiesConfig,
	postgresServerVersionDetector ServerVersionDetector) BackuperFactory {
	return BackuperFactory{
		utilitiesConfig:               utilitiesConfig,
		postgresServerVersionDetector: postgresServerVersionDetector,
	}
}

func (f BackuperFactory) Make(action Action, config config.ConnectionConfig) Interactor {
	switch {
	case config.Adapter == "postgres" && action == "backup":
		return f.makePostgresBackuper(config)
	case config.Adapter == "mysql" && action == "backup":
		return f.makeMysqlBackuper(config)
	case config.Adapter == "postgres" && action == "restore":
		return postgres.NewRestorer(config, f.utilitiesConfig.Postgres_9_4.Restore)
	case config.Adapter == "mysql" && action == "restore":
		return mysql.NewRestorer(config, f.utilitiesConfig.Mysql.Restore)
	}
	return nil
}

func (f BackuperFactory) makeMysqlBackuper(config config.ConnectionConfig) Interactor {
	return NewCompoundMysqlInteractor(
		mysql.NewBackuper(config, f.utilitiesConfig.Mysql.Dump),
		mysql.NewServerVersionDetector(f.utilitiesConfig.Mysql.Client),
		mysql.NewMysqlDumpUtilityVersionDetector(f.utilitiesConfig.Mysql.Dump),
		config)
}

func (f BackuperFactory) makePostgresBackuper(config config.ConnectionConfig) Interactor {
	// TODO: err
	postgresVersion, _ := f.postgresServerVersionDetector.GetVersion(config)

	postgres94Version := version.SemanticVersion{Major: "9", Minor: "4"}
	if postgres94Version.MinorVersionMatches(postgresVersion) {
		return postgres.NewBackuper(config, f.utilitiesConfig.Postgres_9_4.Dump)
	} else {
		return postgres.NewBackuper(config, f.utilitiesConfig.Postgres_9_6.Dump)
	}
}