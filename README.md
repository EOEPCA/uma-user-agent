<div id="top"></div>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]



<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/EOEPCA/uma-user-agent">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a>

  <h3 align="center">UMA User Agent</h3>

  <p align="center">
    Client to implemented the User Managed Access (UMA) flow as a
    server-side nginx `auth_request` agent.
    <br />
    <a href="https://github.com/EOEPCA/uma-user-agent"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/EOEPCA/helm-charts/tree/main/charts/uma-user-agent">Helm Chart</a>
    ·
    <a href="https://github.com/EOEPCA/uma-user-agent/issues">Report Bug</a>
    ·
    <a href="https://github.com/EOEPCA/uma-user-agent/issues">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#background">Background</a></li>
        <li><a href="#http-interface">HTTP Interface</a></li>
        <li><a href="#nginx-configuration">Nginx Configuration</a></li>
        <li>
          <a href="#agent-configuration">Agent Configuration</a>
          <ul>
            <li><a href="#clientyaml">client.yaml</a></li>
            <li><a href="#cconfigyaml">config.yaml</a></li>
          </ul>
        </li>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

### Background

All access attempts to a Resource Server, (e.g. Catalogue, Data Access, ADES, etc.), are subject to the policy enforcement of the PEP (Policy Enforcement Point). For reasons of performance, it is desirable for Nginx to act as the reverse-proxy in this protection flow, rather than the proxy function of the PEP which is not designed for large numbers of concurrent requests or large data volumes.

Through the [Module ngx_http_auth_request_module](https://nginx.org/en/docs/http/ngx_http_auth_request_module.html), Nginx provides a mechanism in which its reverse-proxy function can defer the authorization decision to a subrequest, and so offer protected access. Hence, this `auth_request` interface offers a means to invoke the services of the PEP, whilst maintaining proxy performance.

The `auth_request` approach invokes the subrequest with the expectation to receive one of three possible responses: `2xx (OK)`, `401 (Unauthorized)`, `403 (Forbidden)`. Only 2xx (OK) `auth_request` responses will permit the onward proxy of the request. Otherwise the 401/403 response is returned to the client.

In order to inform its decision the PEP is provided with pertinent request information through http headers set by nginx in the subrequest:
* `X-Original-Method`: http method of client request
* `X-Original-Uri`: path to the requested resource

The PEP implements the nginx `auth_request` interface and so returns 2xx, 401 or 403. In the case of a 401 response, then the PEP expects the client to follow the UMA (User Managed Access) flow, using the 'ticket' that is provided in the `Www-Authenticate` header it returns with the 401 response.

A typical client, such as a browser, is not in a position to follow the UMA flow. Thus, the `uma-user-agent` performs the role of UMA client on behalf of the end-user client (user agent). The uma-user-agent sits between nginx and the PEP, to intercept the PEP 401 responses (with `Www-Authenticate` header) to follow the UMA flow, exchanging a 'ticket' for an RPT (Relying Party Token), which can then be re-presented to the PEP and so gain authorization.

This flow, and the chaining of the uma-user-agent -> PEP in the nginx `auth_request` subrequest, is illustrated in the following sequence diagram.

![Nginx auth_request](uml/export/Nginx%20auth_request.png)

<p align="right">(<a href="#top">back to top</a>)</p>

### HTTP Interface

The uma-user-agent implements the nginx `auth_request` subrequest, as described here http://nginx.org/en/docs/http/ngx_http_auth_request_module.html.

**HTTP Inputs**

The uma-user-agent supports the following inputs on the http subrequest from nginx:
* http headers:
  * `X-Original-Method`: http method of client request
  * `X-Original-Uri`: path to the requested resource
  * `X-User-Id`: user ID token from OIDC (optional)
* http cookie:
  * `auth_user_id`: user ID token from OIDC (optional)<br>
    _Cookie name is configurable_
  * `auth_rpt-<endpoint-name>`: RPT from previous successful access<br>
    _Cookie name is configurable_

**HTTP Outputs**

The uma-user-agent sets the following headers in the http response:
* `2xx (OK)`
  * `X-User-Id`: user ID token, to be passed-on to the target _Resource Server_
  * `X-Auth-Rpt`: RPT from successful authorization
* `401 (Unauthorized)`
  * Www-Authenticate: defines http authorization methods
* `403 (Forbidden)`<br>
  _No specific headers_

<p align="right">(<a href="#top">back to top</a>)</p>

### Nginx Configuration

Nginx must be configured to 1) invoke the `auth_request` subrequest, 2) set the values to be passed in the http subrequest, and 3) to handle any output http headers from the subrequest. For example...

```
  location /resource-server/ {
    auth_request /authcheck;
    auth_request_set $x_user_id $upstream_http_x_user_id;
    auth_request_set $x_auth_rpt $upstream_http_x_auth_rpt;
    proxy_set_header X-User-Id $x_user_id;
    add_header Set-Cookie $x_auth_rpt;
  }

  location ^~ /authcheck {
    internal;
    proxy_pass http://<uma-user-agent-host>/;
    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
    proxy_set_header X-Original-URI $request_uri;
    proxy_set_header X-Original-Method $request_method;
  }
```

<p align="right">(<a href="#top">back to top</a>)</p>

### Agent Configuration

The uma-user-agent reads its configuration from files in the directory specified by the `CONFIG_DIR` environment variable. In the absence of override the default diectory is `/app/config/`.

Configuration is read from two files:
* `client.yaml`<br>
  Details for the client that is registered with the Authorization Server.
* `config.yaml`<br>
  General application configuration.

#### client.yaml

The `client.yaml` file supports the following values:

| Name | Description | Default |
| ---- | ----------- | ------- |
| client-id | The `ID` of the client registered in the Authorization Server | n/a |
| client-secret | The `Secret` of the client registered in the Authorization Server | n/a |

#### config.yaml

The `config.yaml` file supports the following values:

| Name | Description | Default |
| ---- | ----------- | ------- |
| logging.level | Logging level:<br>`panic`, `fatal`, `error`, `warn`/`warning`, `info`, `debug`, `trace` | `info` |
| network.httpTimeout | Timeout for all http client requests (secs) | `10` |
| network.listenPort | Listening port for the uma-user-agent service | `80` |
| pep.url | URL for the PEP, to daisy-chain the `auth_request` call | `http://pep` |
| userIdCookieName | Name of the cookie that carries the User Id Token | `auth_user_id` |
| authRptCookieName | Name of the cookie that carries the RPT of the last successful request<br>Note that this is a prefix for the name that is appended with `-<endpoint-name>` | `auth_rpt` |
| authRptCookieMaxAge | Maximum age of the RPT cookie, to set the expiry (secs) | `300` |
| unauthorizedResponse | Text that should form the value for the `Www-Authenticate` header in the `401` response | n/a |
| retries.authorizationAttempt | Number of retry attempts in the case of an unexpected unauthorized response - i.e. the UMA flow has been successfully followed to obtain a fresh RPT, but it is still rejected<br>A zero `0` value means no retries. | `1` |
| retries.httpRequest | Number of retry attempts in the case of an http request that fails due to specific conditions:<br>* 5xx status code (i.e. server-side error)<br>* Request timeout (i.e. unresponsive server)<br>A zero `0` value means no retries. | `1` |
| openAccess | Boolean to set 'open' access to the resource server.<br>A value of `true` bypasses protections | `false` |

<p align="right">(<a href="#top">back to top</a>)</p>

### Built With

The `uma-user-agent` is implemented using the [Go](https://golang.org/) programming language, with support from the following modules:

* Runtime:
  * [fsnotify](https://github.com/fsnotify/fsnotify) v1.5.1
  * [gorilla/handlers](https://github.com/gorilla/handlers) v1.5.1
  * [gorilla/mux](https://github.com/gorilla/mux) v1.8.0
  * [logrus](https://github.com/sirupsen/logrus) v1.8.1
  * [viper](https://github.com/spf13/viper) v1.8.1
* Build:
  * [air](https://github.com/cosmtrek/air)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

To get a local copy up and running follow these simple example steps.

### Prerequisites

To run natively, the go runtime is required - it can be installed by downloading from here https://golang.org/dl/.

Alternatively, it can be run locally via `docker` and `docker-compose`, which can be installed by downloading from:
* `docker`: https://docs.docker.com/engine/install/
* `docker-compose`: https://docs.docker.com/compose/install/

### Installation

Clone the repo from GitHub...
```sh
git clone https://github.com/EOEPCA/uma-user-agent.git
```

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

```sh
cd uma-user-agent
```

**Run Natively**<br>
To run natively execute the script...
```sh
./run.sh
```

**Run Via Docker**<br>
To run via `docker-compose` execute the script...
```sh
/run-docker.sh
```

**Helm Chart (Kubernetes)**<br>
A helm chart is available for deployment to Kubernetes, here https://github.com/EOEPCA/helm-charts/tree/main/charts/uma-user-agent.

Ensure helm is installed (see https://helm.sh/docs/intro/install/)...
```sh
curl -sfL https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash -s -
```

Add the eoepca helm chart repo...
```sh
helm repo add eoepca https://eoepca.github.io/helm-charts
helm repo update
```

Deploy a helm release `my-uma-agent` to the Kubernetes cluster...
```br
helm install my-uma-agent eoepca/uma-user-agent -f my-values.yaml
```

This will deploy with default values, plus overrides from the file `my-values.yaml`.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the ESA Software Community Licence Permissive. See `LICENSE.txt` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

Your Name - [@eoepca](https://twitter.com/eoepca) - eoepca.systemteam@telespazio.com

Project Link: [https://github.com/EOEPCA/uma-user-agent](https://github.com/EOEPCA/uma-user-agent)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* README.md is based on [this template](https://github.com/othneildrew/Best-README-Template) by [Othneil Drew](https://github.com/othneildrew).

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/EOEPCA/uma-user-agent.svg?style=for-the-badge
[contributors-url]: https://github.com/EOEPCA/uma-user-agent/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/EOEPCA/uma-user-agent.svg?style=for-the-badge
[forks-url]: https://github.com/EOEPCA/uma-user-agent/network/members
[stars-shield]: https://img.shields.io/github/stars/EOEPCA/uma-user-agent.svg?style=for-the-badge
[stars-url]: https://github.com/EOEPCA/uma-user-agent/stargazers
[issues-shield]: https://img.shields.io/github/issues/EOEPCA/uma-user-agent.svg?style=for-the-badge
[issues-url]: https://github.com/EOEPCA/uma-user-agent/issues
[license-shield]: https://img.shields.io/github/license/EOEPCA/uma-user-agent.svg?style=for-the-badge
[license-url]: https://github.com/EOEPCA/uma-user-agent/blob/master/LICENSE.txt
