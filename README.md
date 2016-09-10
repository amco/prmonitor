# prmonitor

**prmonitor** is a dedicated dashboard that monitors outstanding
pull request across multiple repositories. It is intended to be
run on displays without any inputs. Any PRs that take longer than
a day are considered risky and get flagged orange. Any longer
than 3 days and they get flagged red.

![Example](/example.png)

## Installation
 1. You'll need godeps if you don't already have it:
    ```
    go get github.com/tools/godep
    ```

 2. Fetch the project and install with godeps.
    ```
    go get github.com/brentdrich/prmonitor
    godep go install
    ```

## Running Locally
 1. Create a [personal access token](https://github.com/blog/1509-personal-api-tokens) at Github

 2. Update the CONFIG variable below and run:
    ```
    PORT=8080 CONFIG='{"dashboard_user":<username>,"dashboard_pass":<password>,"github_user":<your-username>,"github_pass":<your-personal-access-token>,"repos":[{"owner":"docker","repo":"swarmkit","depth":15}]}' prmonitor
    ```

 4. Navigate to `0.0.0.0:8080`

## Running on Heroku
Deploying prmonitor to heroku requires that the vendor directory is included in
the repository. `deploy.sh` is a script that creates a new branch that includes the
vendor directory (so master doesn't have to) and pushes it out to heroku automatically.

 1. Set up a heroku app and define a [CONFIG variable](https://devcenter.heroku.com/articles/config-vars)

 2. Set up [heroku-cli](https://devcenter.heroku.com/articles/deploying-go)

 3. Run the deployment script
    ```
    deploy.sh
    ```