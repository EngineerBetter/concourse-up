package bosh

const concourseManifestFilename = "concourse.yml"
const credsFilename = "concourse-creds.yml"
const concourseDeploymentName = "concourse"
const concourseVersionsFilename = "versions.json"
const concourseSHAsFilename = "shas.json"
const concourseGrafanaFilename = "grafana_dashboard.yml"
const concourseCompatibilityFilename = "cup_compatibility.yml"
const concourseGitHubAuthFilename = "github-auth.yml"
const extraTagsFilename = "extra_tags.yml"

//go:generate go-bindata -pkg $GOPACKAGE -ignore \.git assets/... ../../concourse-up-ops/...
var awsConcourseGrafana = MustAsset("assets/grafana_dashboard.yml")
var awsConcourseCompatibility = MustAsset("assets/ops/cup_compatibility.yml")
var awsConcourseGitHubAuth = MustAsset("assets/ops/github-auth.yml")
var extraTags = MustAsset("assets/ops/extra_tags.yml")
var awsConcourseManifest = MustAsset("../../concourse-up-ops/manifest.yml")
var awsConcourseVersions = MustAsset("../../concourse-up-ops/ops/versions.json")
var awsConcourseSHAs = MustAsset("../../concourse-up-ops/ops/shas.json")
