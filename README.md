# prmonitor

**prmonitor** is a dedicated dashboard that monitors outstanding
pull request across multiple repositories. It is intended to be
run on displays without any inputs.

![example.png]

## Installation
```
go get -u github.com/brentdrich/prmonitor
```

## Usage
 1. Create a [personal access token](https://github.com/blog/1509-personal-api-tokens) at Github

 2. Create a config.yaml file in a directory somewhere
    ```
    username: <your-user-name>
    password: <your-personal-access-token>
    repos:
     - owner: docker
       repo: swarmkit
       depth: 100
    ```

 3. From the same directory, run:
    ```
    prmonitor
    ```

 4. Navigate to 0.0.0.0:8080