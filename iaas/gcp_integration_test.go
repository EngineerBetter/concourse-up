// +build integration

package iaas_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/EngineerBetter/concourse-up/iaas"
	"github.com/EngineerBetter/concourse-up/util"
)

func TestGCPProvider_CreateDatabases(t *testing.T) {
	dbName := util.EightRandomLetters()
	setup(dbName, t)
	defer cleanup(dbName)

	_, passwordErr, err := runCommand("gcloud", "sql", "users", "set-password", "postgres", "--instance", dbName, "--password", "password")
	failIfErr(t, err, "Unable to set password with err [%v] and stderr [%v]", passwordErr)

	gcp, err := iaas.New("GCP", "europe-west1")
	failIfErr(t, err, "Unable to create GCP instance: %v")

	err = gcp.CreateDatabases(dbName, "postgres", "password")
	failIfErr(t, err, "Unexpected error setting database password: %v")

	instancesOut, instancesErr, err := runCommand("gcloud", "sql", "databases", "list", "--instance", dbName, "--format", "json")
	failIfErr(t, err, "Unable to list databases due to err: %v, stderr: %v", instancesErr)

	var databases []map[string]interface{}
	json.Unmarshal([]byte(instancesOut), &databases)

	assertDatabaseExists("concourse_atc", databases, t)
	assertDatabaseExists("credhub", databases, t)
	assertDatabaseExists("uaa", databases, t)
}

func failIfErr(t *testing.T, err error, pattern string, args ...interface{}) {
	errAndArgs := append([]interface{}{err}, args...)
	if err != nil {
		t.Fatalf(pattern, errAndArgs...)
	}
}

func runCommand(program string, args ...string) (string, string, error) {
	cmd := exec.Command(program, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	return outStr, errStr, err
}

func assertDatabaseExists(desired string, databases []map[string]interface{}, t *testing.T) {
	for _, database := range databases {
		if database["name"] == desired {
			return
		}
	}
	t.Fatalf("Did not find database %v in %+v", desired, databases)
}

func setup(dbName string, t *testing.T) {
	_, errStr, err := runCommand("gcloud", "sql", "instances", "create", dbName, "--tier=db-f1-micro", "--database-version=POSTGRES_9_6", "--region=europe-west1")
	re := `ERROR: \(gcloud\.sql\.instances\.create\).*is taking longer than expected. You can continue waiting for the operation by running`
	r, _ := regexp.Compile(re)
	match := r.MatchString(errStr)
	if err != nil && !match { // and also not a message about timing out
		t.Errorf("Cannot set up database for test error = %v", err)
	}

	if match {
		opStr := `([-a-z0-9]*)[^-a-z0-9]*$`
		opRe, _ := regexp.Compile(opStr)
		op := opRe.FindAllStringSubmatch(errStr, -1)
		opName := op[0][1]
		_, _, err = runCommand("gcloud", "beta", "sql", "operations", "wait", "--project=concourse-up", opName)
		failIfErr(t, err, "Failed with err %v  waiting for the operation %v to finish", op)
	}
}

func cleanup(dbName string) {
	if _, skipTeardown := os.LookupEnv("SKIP_TEARDOWN"); skipTeardown {
		fmt.Println("Skipping teardown")
		return
	}

	runCommand("gcloud", "sql", "instances", "delete", dbName, "--async", "--quiet")
}
