readonly listCheckRuns=$(curl -L \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN"\
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/$OWNER/$REPO/commits/$REF/check-runs")


readonly jobID=$(echo "$listCheckRuns" |
    jq --arg name "$JOB_TO_RERUN" '.check_runs[] | select(.name == $name) | .id')

echo "rerun jovb $jobID"

curl -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_TOKEN"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/$OWNER/$REPO/actions/jobs/$jobID/rerun