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

## Usage
 1. Create a [personal access token](https://github.com/blog/1509-personal-api-tokens) at Github

 3. Update the CONFIG variable below and run:
    ```
    PORT=8080 CONFIG='{"username":<your-username>,"password":<your-personal-access-token>,"repos":[{"owner":"docker","repo":"swarmkit","depth":15}]}' prmonitor
    ```

 4. Navigate to `0.0.0.0:8080`