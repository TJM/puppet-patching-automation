# DevOps SRE - Patching Automation

Go application for the patching automation project.

## Install

* Basic Helm installation

```bash
helm repo add BLAH URLTBD
helm repo update
helm upgrade --install patching-automation BLAH/patching-automation
```

## Upgrade

### v0.12.x to v0.13.x

* Updated `bitnami/postgresql` chart from 9.x to 11.x.
Please see *their* [upgrade docs](https://docs.bitnami.com/kubernetes/infrastructure/postgresql/administration/upgrade/).
Some of the authentication and persistence parameters are now deprecated as they have been moved under `primary`.
We have had pretty good luck with just removing the helmRelease and re-installing it after getting our values in order.
This does cause a small downtime, but re-starts with all the old data.

## Building (compiling) patching-automation

### Local Build Prequisites

* Go 1.20+
* make

### Building Locally

* Full build process (linting, gosec, etc.):

```bash
make
```

* Build the application only:

```bash
make patching-automation
```

## Running Locally (for testing)

After compiling above, here are the basics:

```bash
./patching-automation --version
./patching-automation --help
```

### Local Run Prerequisites

Required:

* Authentication (OIDC) Configuration: (See [Authentication Options](#authentication-options) below)
  * OIDC_CLIENT_ID=xxx.xxxx.xxxxxx
  * OIDC_CLIENT_SECRET=xxxxxxxxx
  * OIDC_REDIRECT_URL="http://127.0.0.1:8080/auth/google/callback"
  * OIDC_ISSUER_URL="https://oidc.example.com"

* `PuppetServer` - This application will *query* PuppetDB, so you need access to a PuppetServer to actually use the application, once it is running.
  * Visit: `http://localhost:8080/config/puppetServer` to add a PuppetServer

### Get Puppet Server token

* Get Puppet CA Cert from: `cat /etc/puppetlabs/puppet/ssl/certs/ca.pem` (from any puppet client or the puppet primary)
  * Be sure to include the **BEGIN** and **END** lines

* Create a Patching Automation service account in Puppet with the following access:
  * Nodes -> View node data from PuppetDB
  * Job Orchestrator -> Start, Stop and view jobs
  * Plans -> Run Plans ->
    * `patchy::cluster_patching`
    * `patchy::patching`
  * Tasks -> Run Tasks ->
    * `pe_patch::agent_health`
    * `pe_patch::clean_cache`
    * `pe_patch::last_boot_time`
    * `pe_patch::patch_server`
    * `pe_patch::refresh_fact`

* To get a puppet token, SSH into the puppet server and run:
  * `puppet access login --username SERVICEACCOUNT --print --lifetime 30d`
    * Adjust the lifetime option for the token to suit your needs.

Optional:

* `JenkinsServer` - If you want to be able to run Jenkins jobs, you will need a Jenkins Server.
  * Visit `http://localhost:8080/config/jenkinsServer` to add a JenkinsServer
  * For local development, you can use `docker run -p 8000:8080 -v jenkins_home:/var/jenkins_home jenkins/jenkins:lts` (and add localhost:8000 as a jenkinsServer).
* `TRELLO_TOKEN` - If you want to generate a trello board, you will need a trello token. This grants this specific application (patching-automation) full access to *your* trello account. You can get this by visiting the following URL: <https://trello.com/1/connect?key=5a453a8d5b4ab0ae9a5746b34cc0b09e&name=PatchingAutomation&response_type=token&scope=read,write>

* See Database Options - By default local development will use a local sqlite3 DB, which is fine for local development, but other options are available. See Database Options below...

### Starting Server

While developing, there is a shortcut `make run` which will compile *and* start the server. This is extremely handy when you are actively developing and want to test changes. NOTE: if you are only changing templates (views), you do not need to recompile, as they will be picked up automatically.

```bash
./patching-automation
```

NOTE: If you receive an error about port 8080 being in use, you can select a different port with `export PORT=8081` for example before running. You will need to adjust each of the examples below if you change the port.

## Interacting with the WebUI

* Visit the URL to the server, by default: `http://localhost:8080/`
* Click the `[Patch Runs]` button to enter the main part of the application.
* There are several different interfaces...
  * Puppet
    * Puppet Servers (`/config/puppetServer`) - Add/Manage Puppet Servers
    * Puppet Tasks (`/config/puppetTask`) - Add/Manage Puppet Tasks (on Puppet Servers)
    * Puppet Plans (`/config/puppetPlan`) - Add/Manage Puppet Plans (on Puppet Servers)
  * Jenkins Servers (`/config/jenkinsServer`) - Add/Manage Jenkins Servers
    * Jenkins Jobs (`/config/jenkinsJob`) - Add/Manage JenkinsJobs/Params (inside Jenkins Servers)
  * Chat Rooms (`/config/ChatRoom`) - Add/Manage Chat Rooms (WebEx Teams) for notifications
  * Patch Runs (`/patchRun`) - Add/Manange Patch Runs (the main point of the app)
    * /application, /environment, /component, /server - view information about sub-parts of patch runs.

## Interacting with the API

NOTE: Currently the JSON API is **broken**. When we enabled Authentication, the API interactions became more difficult. We have an open issue to figure out and document how to authenticate with the API in the future.

You can interact with the JSON API using a tool like [Postman](https://www.postman.com/downloads/), or you can use curl. The paths are the same as above, but with a header like: `Accept: application/json`

You can also use `applcation/x-yaml` or `application/xml` if you prefer.

-----

## Authentication Options

We currently use [OpenID Connect (OIDC)](https://en.wikipedia.org/wiki/OpenID_Connect) for authentication.

* Oauth2 Identity Provider (IdP) service that supports [OIDC](https://en.wikipedia.org/wiki/OpenID_Connect)
  * You can use something like [DEX](https://github.com/dexidp/dex) to test with.
  * Alternatively, you could also use something like:
    * [Google](https://developers.google.com/identity/protocols/oauth2/openid-connect)
    * [GitHub](https://plugins.miniorange.com/oauth-openid-login-using-github)
    * etc

### DEX Identity Provider

The example below will use [DEX IdP](https://dexidp.io/). Please clone their [repo](https://github.com/dexidp/dex) and start DEX in a separate window. You can follow their [Getting Started](https://dexidp.io/docs/getting-started/#building-the-dex-binary) docs.

* Start DEX IdP:

NOTE: This is only for example purposes, and should not be used in production. You can configure dex properly to authenticate against corporate LDAP or many other sources.

```console
./bin/dex serve examples/config-dev.yaml
```

* Create a `.env` file with the following contents:

```bash
PORT=5555
OIDC_CLIENT_ID="example-app"
OIDC_CLIENT_SECRET="ZXhhbXBsZS1hcHAtc2VjcmV0"
OIDC_REDIRECT_URL="http://127.0.0.1:5555/callback"
OIDC_ISSUER_URL="http://127.0.0.1:5556/dex"
INIT_ADMINS="admin@example.com"
INIT_USERS="kilgore@kilgore.trout"
```

NOTE: Since we are using the dex example configuration, you will need to access Patching Automation as <http://127.0.0.1:5555/>

### Google Accounts Identity Provider

The example below will use Google Accounts. See: [go-oidc examples readme](https://github.com/coreos/go-oidc/tree/v3/example).

* Setup Google

  1. Visit your [Google Developer Console](https://console.developers.google.com/).
  1. **ONLY DO ONCE**: Click "Configure Consent Screen"
        1. Select **Internal**
        2. Create
        3. At a minimum fill in a name for your App: ie: local_app, user support email, and Developer email.
  1. Click "Credentials" on the left column.
  1. Click the "Create credentials" button followed by "OAuth client ID".
  1. **Application type** - "Web application"
  1. **Name** - Add a unique Name.
  1. **Authorized JavaScript origins** - Skip
  1. **Authorized redirect URIs**
      1. *Add URI*: "http://127.0.0.1:8080/auth/google/callback"
  1. Click create and add the printed client ID and secret to your environment using the following variables: (`.env` file)

```bash
OIDC_CLIENT_ID="(value provided by google)"
OIDC_CLIENT_SECRET="(value provided by google)"
OIDC_REDIRECT_URL="http://127.0.0.1:8080/auth/google/callback"
OIDC_ISSUER_URL="https://accounts.google.com"
INIT_ADMINS="your.email@company.com"
```

-----

## Database Options

By default, the application will create a local sqlite3 database called `db/padb.db` which is sufficient for local development. These other database options are also supported: mysql, postgresql

### MySQL Example

NOTE: This will persist the database data in `./db/mysql`, if you need to start over clean or just want to clean up, make sure the database is stopped, then you can remove the directory.

* Start the Database

```bash
docker run -p 3306:3306 -d --rm -P -e MYSQL_DATABASE=padb -e MYSQL_USER=padb -e MYSQL_PASSWORD=padb -e MYSQL_ROOT_PASSWORD=notEmpty123 -v $PWD/db/mysql:/bitnami/mysql/data --name padb-mysql bitnami/mysql
```

NOTE: Port is hard coded to 3306, if you change it you will have to add `--dbport xxxx` below

* Run patching-automation with requisite options:

```bash
 ./patching-automation --dbtype mysql
```

* Stop the database (when you are done with it)

```bash
docker stop padb-mysql
```

### PostgreSQL Example

NOTE: This will persist the database data in `./db/postgresql`, if you need to start over clean or just want to clean up, make sure the database is stopped, then you can remove the directory.

* Start the database

```bash
docker run -p 5432:5432 -d --rm -v $PWD/db/postgresql:/bitnami/postgresql -P -e POSTGRES_USER=padb -e POSTGRES_PASSWORD=padb -e POSTGRES_DATABASE=padb --name padb-postgresql bitnami/postgresql
```

NOTE: Port is hard coded to 5432, if you change it you will have to add `--dbport xxxx` below

* Run patching-automation with requisite options:

```bash
 ./patching-automation --dbtype postgresql
```

* Stop the database (when you are done with it)

```bash
docker stop padb-postgresql
```

-----

## Docker

### Building Docker image

* Build Docker Image

```bash
make docker
```

... do whatever testing you deem fit.

### Publish Docker image

```bash
make dockerpush
```

## Running in Docker

### Prerequisites

* Database: For local development, we can use sqlite3, but docker images are ephemeral. Assuming you want to keep your state, you probably want to use an external database. We will create a test database below in a "user defined network" so they can find eachother.

### Running

Basic Test:

```bash
docker run patching-automation --version
docker run patching-automation --help
```

NOTE: You will need a database (as per Database Options). The important option below is `--network pa-test`.

* Create a docker Network

```bash
docker network create pa-test
```

* Start a database (postgresql in this example, see options [above](#database-options))

```bash
docker run -d --rm --network pa-test -v $PWD/db/postgresql:/bitnami/postgresql  -e POSTGRES_USER=padb -e POSTGRES_PASSWORD=padb -e POSTGRES_DATABASE=padb --name padb-postgresql bitnami/postgresql
```

* Start Application

```bash
docker run -p 8080:8080 --rm --network pa-test patching-automation --dbtype postgresql --dbhost padb-postgresql
```

NOTE: We are hard-coding the port to 8080, you can change as needed.

ALSO: The database was created with the "default" values the application expects. See `--help` option for more details

-----

## Helm Chart

We also have a Helm Chart to deploy patching-automation to Kubernetes.

### Building Helm Chart

NOTE: The steps below requires [helm](https://helm.sh/docs/intro/install/) v3.x to be installed locally.

```bash
make helm
```

### Testing Helm Chart

Getting access to a Kubernetes cluster or creating your own through kind/talos/whatever is outside the scope of this document.

* Test rendering templates locally

```bash
helm template helmchart/patching-automation
```

* Install test version

```bash
helm upgrade --install ./helmchart/patching-automation -f values.yaml
```

### Publish Helm Chart

We publish our helm chart to Artfactory, see Makefile variables for details.

```bash
make helmpush
```

### Install using helm chart

We usually use a helmRelease in flux, but to install the chart, you will need a values file.

* Create `values.yaml`

```yaml
# Values for example deployment of patching-automation

## Ingress
ingress:
  enabled: true
  annotations:
    kubernetes.io/ingress.class: nginx
    # probably other options here?
  hosts:
    - host: pa.example.com
      paths: ["/"]
  tls:
    - hosts:
        - pa.example.com

## Database
postgresql:
  postgresqlPassword: sUp3rS3cr3tDatabasePW

## Environment Variables to Configure Patching Automation
extraEnvironmentVariables:
  - name: TRELLO_TOKEN
    value: f22XXXXXXXXXXXXXXXXX
  - name: OIDC_CLIENT_ID
    value: 4XXXXXXXXXr.apps.googleusercontent.com
  - name: OIDC_CLIENT_SECRET
    value: GOXXXXXXXXXXXXe5
  - name: OIDC_REDIRECT_URL
    value: https://pa.example.com/auth/google/callback
  - name: OIDC_ISSUER_URL
    value: https://accounts.google.com

```

* Install/Upgrade

```bash
helm upgrade --install patching-automation REPO/patching-automation -f values.yaml
```

-----

## Release a new version

Releasing a new version has been automated through the Makefile. Running "`make release VERSION=x.y.z`" will create a new branch, update the version files and create a merge request.

* Run `make release` with a VERSION variable
  * Check current tag release. Two ways.
    * Check `const Version` variable in `version/version.go`.
    * Check gitlab tags in repository.
  * We are using semantic versioning. (MAJOR.MINOR.PATCH)
  * Do *not* insert the **v** in the VERSION variable below.
  * **OPTIONAL**: If you want to create a pre-release, such as `rc1`, `beta`, or whatever, also set "PRELEASE_TAG" (otherwise leave it *empty*) (`make release VERSION=0.12.0 PRERELEASE_TAG=beta`)
  * **EDIT THIS ONE!**

```bash
make release VERSION=X.Y.Z
```

* Get approvals and merge branch in GitLab UI.

* TAG the release on the **master** branch *after* merging. -- NOTE: This requires "maintainer" access. Command line steps below or you can compelete this step in the GitLab UI.

```bash
git checkout master && git pull
git tag v${VERSION}
git push origin v${VERSION}
```

### Post-Release Steps

* Run `make prepare_for_dev` with VERSION variable
  * We are using semantic versioning. (MAJOR.MINOR.PATCH)
  * Set this to the next "minor" version for development. If any "patches" (hotfixes)  are needed, they will be the PATCH release.
    * Example: If the current release is `0.11.2`, the next dev release will be `0.11.3`
    * Previously we set it to the next minor release, but this causes issues when auto-releasing helm charts.
  * The `PRERELEASE_TAG` will be set to `dev`
  * Do *NOT* insert the **v** in the VERSION variable below.
  * **EDIT THIS ONE!**

```bash
make prepare_for_dev VERSION=X.Y.Z
```

* Get approvals and merge branch in GitLab UI.
