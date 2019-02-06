package testsupport

import (
	"io/ioutil"
	"os"
	"testing"
)

// CompareActions compares strings a and b in slice actions
// and returns:
// >0 if a.index > b.index
// <0 if a.index < b.index
// 0 if a.index == b.index
func CompareActions(actions []string, a, b string) int {
	m := make(map[string]int)
	for i, e := range actions {
		m[e] = i
	}
	return m[a] - m[b]
}

func SetupFakeCredsForGCPProvider(t *testing.T) string {
	json := `{"project_id": "fake_id", "type": "service_account"}`
	filePath, err := ioutil.TempFile("", "")
	if err != nil {
		t.Error("Could not create GCP credentials file")
	}
	_, err = filePath.WriteString(json)
	if err != nil {
		t.Error("Could not write in GCP credentials file")
	}
	filePath.Close()
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", filePath.Name())
	if err != nil {
		t.Errorf("cannot set %v", err)
	}
	return filePath.Name()
}
