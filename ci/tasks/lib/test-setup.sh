#!/bin/bash
# shellcheck disable=SC1091

# Disabling SC1091 because shellcheck can't find our sourced files

source concourse-up/ci/tasks/lib/set-flags.sh
source concourse-up/ci/tasks/lib/assert-iaas.sh
source concourse-up/ci/tasks/lib/verbose.sh
source concourse-up/ci/tasks/lib/id.sh
source concourse-up/ci/tasks/lib/pipeline.sh
source concourse-up/ci/tasks/lib/trap.sh
source concourse-up/ci/tasks/lib/credhub.sh
