SHA=$(curl -L \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN"\
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/$OWNER/$REPO/pulls/$PR_NUMBER" | \
    jq -r '.head.sha')

echo "SHA of the last commit in the PR: $SHA"

readonly listCheckRuns=$(curl -L \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GITHUB_TOKEN"\
    -H "X-GitHub-Api-Version: 2022-11-28" \
    "https://api.github.com/repos/$OWNER/$REPO/commits/$SHA/check-runs")


readonly jobID=$(echo "$listCheckRuns" |
    jq --arg name "$JOB_TO_RERUN" '.check_runs[] | select(.name == $name) | .id')


if [ -z "$jobID" ]; then
    echo "No job with name $JOB_TO_RERUN found in the last commit of the PR"
    exit 1
fi

echo "rerun job $jobID"

curl -L \
  -X POST \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_TOKEN"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/$OWNER/$REPO/actions/jobs/$jobID/rerun
