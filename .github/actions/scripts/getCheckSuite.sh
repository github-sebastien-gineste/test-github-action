OWNER="ebuzzing"
REPO="service-api-domains"
PR_NUMBER=8790
GITHUB_TOKEN=
REF="49e8443f0df74798ece63cbaab909b6b676fbdb0"
CHECK_SUITE_ID=13563876255

curl -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer $GITHUB_TOKEN"\
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "https://api.github.com/repos/$OWNER/$REPO/commits/$REF/check-suites"| \
    jq -r '.check_suites[] | [.id, .app.name, .status, .conclusion]'

#curl -L \
#  -H "Accept: application/vnd.github+json" \
#  -H "Authorization: Bearer $GITHUB_TOKEN"\
#  -H "X-GitHub-Api-Version: 2022-11-28" \
#  https://api.github.com/repos/$OWNER/$REPO/check-suites/$CHECK_SUITE_ID

#curl -L \
#  -H "Accept: application/vnd.github+json" \
#  -H "Authorization: Bearer $GITHUB_TOKEN"\
#  -H "X-GitHub-Api-Version: 2022-11-28" \
#  "https://api.github.com/repos/$OWNER/$REPO/check-suites/$CHECK_SUITE_ID/check-runs"