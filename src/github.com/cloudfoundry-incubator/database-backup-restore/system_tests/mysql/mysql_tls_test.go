package mysql

import (
	"fmt"
	"net"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/go-sql-driver/mysql"

	"strings"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/system_tests/utils"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql with tls", func() {
	if os.Getenv("TEST_TLS") == "false" {
		fmt.Println("**********************************************")
		fmt.Println("Not testing TLS")
		fmt.Println("**********************************************")
		return
	}
	var dbDumpPath string
	var configPath string
	var databaseName string
	var mysqlSslUsername string
	var mysqlMutualTlsUsername string
	var configJson string

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString

		RunSQLCommand("CREATE DATABASE "+databaseName, connection)
		RunSQLCommand("USE "+databaseName, connection)
		RunSQLCommand("CREATE TABLE people (name varchar(255));", connection)
		RunSQLCommand("INSERT INTO people VALUES ('Old Person');", connection)

		mysqlSslUsername = "ssl_user_" + DisambiguationStringOfLength(6)
		RunSQLCommand(fmt.Sprintf(
			"CREATE USER '%s' IDENTIFIED BY '%s';",
			mysqlSslUsername, mysqlPassword), connection)

		mysqlMutualTlsUsername = "mtls_user_" + DisambiguationStringOfLength(6)
		RunSQLCommand(fmt.Sprintf(
			"CREATE USER '%s' IDENTIFIED BY '%s';",
			mysqlMutualTlsUsername, mysqlPassword), connection)
	})

	AfterEach(func() {
		RunSQLCommand(fmt.Sprintf("DROP USER '%s';", mysqlSslUsername), connection)
		RunSQLCommand(fmt.Sprintf("DROP USER '%s';", mysqlMutualTlsUsername), connection)
	})

	JustBeforeEach(func() {
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	Context("when the db user requires TLS", func() {
		BeforeEach(func() {
			RunSQLCommand(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO %s REQUIRE SSL;", databaseName, mysqlSslUsername), connection)
		})

		Context("when TLS info is not provided in the config", func() {
			BeforeEach(func() {
				configJson = fmt.Sprintf(
					`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "mysql"
					}`,
					mysqlSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
			})

			It("works", func() {
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
						dbDumpPath, configPath))

				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
						dbDumpPath, configPath))

				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
			})
		})

		Context("when TLS info is provided in the config", func() {
			Context("and host verification is not skipped", func() {
				if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
					fmt.Println("**********************************************")
					fmt.Println("Not testing TLS with Verify Identity")
					fmt.Println("**********************************************")
					return
				}
				Context("and the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
					})
				})
				Context("and the host does not match the CA cert", func() {
					BeforeEach(func() {
						By("connecting to the host using its IP rather than the hostname")
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlSslUsername,
							mysqlPassword,
							resolveHostToIP(mysqlHostName),
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("fails as the hosts does not match the certificate", func() {
						Expect(brJob.RunOnInstance(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))).To(gexec.Exit(1))

					})
				})
			})
			Context("and host verification is skipped", func() {
				Context("and the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"skip_host_verify": true,
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
					})
				})
				Context("and the CA cert is malformed", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "mysql",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "fooooooo"
										}
									}
								}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})
			})
		})
	})

	Context("when the db user requires TLS and Mutual TLS", func() {
		if os.Getenv("TEST_TLS_MUTUAL_TLS") == "false" {
			fmt.Println("**********************************************")
			fmt.Println("Not testing TLS with Mutual TLS")
			fmt.Println("**********************************************")
			return
		}
		BeforeEach(func() {
			RunSQLCommand(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO %s REQUIRE SSL;", databaseName, mysqlMutualTlsUsername), connection)
			RunSQLCommand(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO %s REQUIRE X509;", databaseName, mysqlMutualTlsUsername), connection)
		})

		Context("when TLS info is not provided in the config", func() {
			BeforeEach(func() {
				configJson = fmt.Sprintf(
					`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "mysql"
					}`,
					mysqlMutualTlsUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
			})

			It("does not work", func() {
				Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
					"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
			})
		})

		Context("when TLS info is provided in the config", func() {
			Context("and host verification is not skipped", func() {
				if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
					return
				}
				Context("and the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"cert": {
									"ca": "%s",
									"certificate": "%s",
									"private_key": "%s"
								}
							}
						}`,
							mysqlMutualTlsUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
							escapeNewLines(mysqlClientCert),
							escapeNewLines(mysqlClientKey),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
					})
				})
			})

			Context("and host verification is skipped", func() {
				Context("and the client cert and key are provided and correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "mysql",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "%s",
											"certificate": "%s",
											"private_key": "%s"
										}
									}
								}`,
							mysqlMutualTlsUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
							escapeNewLines(mysqlClientCert),
							escapeNewLines(mysqlClientKey),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
					})
				})

				Context("and the client cert and key are provided and malformed", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "mysql",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "%s",
											"certificate": "foo",
											"private_key": "bar"
										}
									}
								}`,
							mysqlMutualTlsUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})

				Context("and the client cert and key are not provided", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"skip_host_verify": true,
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlMutualTlsUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})
			})
		})
	})

	Context("when the db user does not require TLS", func() {
		Context("and host verification is not skipped", func() {
			if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
				return
			}
			Context("and the CA cert is correct", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "mysql",
						"tls": {
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
						databaseName,
						escapeNewLines(mysqlCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
				})
			})
		})

		Context("and host verification is skipped", func() {
			Context("and the CA cert is correct", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "mysql",
						"tls": {
							"skip_host_verify": true,
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
						databaseName,
						escapeNewLines(mysqlCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
				})
			})
			Context("and the CA cert is malformed", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
								"username": "%s",
								"password": "%s",
								"host": "%s",
								"port": %d,
								"database": "%s",
								"adapter": "mysql",
								"tls": {
									"skip_host_verify": true,
									"cert": {
										"ca": "fooooooo"
									}
								}
							}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
						databaseName,
					)
				})

				It("does not work", func() {
					Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
						"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
				})
			})
		})
	})
})

func resolveHostToIP(hostname string) string {
	addrs, err := net.LookupHost(hostname)
	Expect(err).NotTo(HaveOccurred())
	Expect(addrs).NotTo(HaveLen(0), "hostname "+hostname+" does not resolve to any IPs")
	return addrs[0]
}

func escapeNewLines(txt string) string {
	return strings.Replace(txt, "\n", "\\n", -1)
}
