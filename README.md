# prmonitor

**prmonitor** is a dedicated dashboard that monitors pull requests
for a team across multiple repositories. Pull requests are displayed
on a reverse [Gantt Chart](https://en.wikipedia.org/wiki/Gantt_chart) running
from present (right) to past (left), with bar length as review duration
and bar position as the review time.

![Example](/example.png)

Features:

 * See a snapshot of the depth and duration of outstanding
   pull requests. (depth - how many are out at a given time,
   duration - how long is a pull request under review).

 * Measure improvements by looking at the historical pull
   requests displayed below the open pull requests.

 * See the pull request activity across multiple repositories,
   which is useful for organizations going towards microservices.

 * Filter pull requests to only see the ones by team members,
   which is useful for organizations where many teams are committing
   to a single repository.

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

    If the above fails due to missing deps, please run:
    ```
    godep restore
    ```

## Development
The easiest way to hack on the dashboard is to run:

```
go test && open e2e.html
```

## Deployment to Heroku
Deploying prmonitor to heroku requires that the vendor directory is included in
the repository. `deploy.sh` is a script that creates a new branch that includes the
vendor directory (so master doesn't have to) and pushes it out to heroku automatically.

* Set up a heroku app and define [environment variables](https://devcenter.heroku.com/articles/config-vars). Examples:
  * CONFIG={"repos": [{"owner": "brentdrich", "repo": "prmonitor"}],
            "customization": {"passiveColor": "#00cc66", "warningColor": "#ffff00",
                              "alertColor":   "#cc0000", "passiveTime":  24.0,
                              "closedColor": "#999", "warningTime":  48}
            }
  * SSLHOST=myapp.herokuapp.com
  * PORT=8081
  * DASHBOARD_USER=admin
  * DASHBOARD_PASSWORD=admin
  * GITHUB_TOKEN=xxx
* Set up [heroku-cli](https://devcenter.heroku.com/articles/deploying-go)
* Run the deployment script
    ```
    deploy.sh
    ```

The dashboard can be run on a free hobby dyno and refreshes once every 24 hours.

## Customized ranges and colors
    # Add customizations to the CONFIG env value as shown in the example for settingup heroku.
    # Add values for passive time and color, and warning time and color, and alert color
    # This will simply change at what points the bars change color based on how quickly they
        are closed.
