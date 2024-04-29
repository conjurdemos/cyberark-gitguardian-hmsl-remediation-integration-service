<img alt="CyberArk Banner" src="images/cyberark-banner.jpg">

<!--
Author:   David Hisel <david.hisel@cyberark.com>
Updated:  <2024/03/19 02:08:05>
-->

# CyberArk GitGuardian Integration Guide

## Revision History

| Revision Number | Author      | Summary                               |
|-----------------|-------------|---------------------------------------|
| 20240401-01     | David Hisel | Version 1.0                           |
| 20240418-01     | David Hisel | Version 1.1 – Add Credential Provider |

## Table Of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

* [General Requirements](#general-requirements)
* [Linux Host Configuration](#linux-host-configuration)
  * [Linux Host Requirements](#linux-host-requirements)
  * [Setup The Frontend Proxy](#setup-the-frontend-proxy)
* [GitGuardian Configuration](#gitguardian-configuration)
  * [GG Requirements](#gg-requirements)
  * [Generate Personal Access Token](#generate-personal-access-token)
* [Create the GitGuardian Webhook](#create-the-gitguardian-webhook)
* [Install and Build the Integration Service](#install-and-build-the-integration-service)
  * [Configure The Integration Service](#configure-the-integration-service)
  * [Create a startup script](#create-a-startup-script)
* [Credential Provider Installation](#credential-provider-installation)
  * [Credential Provider Deamon Installation](#credential-provider-deamon-installation)
  * [Credential Provider SDK Installation](#credential-provider-sdk-installation)
* [Install and Build the Integration Service with Credential Provider](#install-and-build-the-integration-service-with-credential-provider)
  * [Install the CyberArk GitGuardian Integration Host Platform](#install-the-cyberark-gitguardian-integration-host-platform)
* [Configure The CP Integration Service](#configure-the-cp-integration-service)
  * [Create a Privilege Cloud Application](#create-a-privilege-cloud-application)
* [Create a CP startup script](#create-a-cp-startup-script)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## General Requirements

Customers must have the following in order to implement this capability:

* Privilege Cloud Account
  * Safe created for storing Pending accounts
  * User that has access to the Pending safe with sufficient permissions
    to add accounts to the safe
  * CPM Plugin installed – CPM plugin can be downloaded from the CyberArk
    marketplace

* Linux Host – this is where the integration application will run

* PostgreSQL (or PostgreSQL compatible) database accessible by the Linux
  Machine

* Gitguardian Account
  * Personal Access Token
  * Webhook Authorization Token

To enable the **Credential Provider** feature, the customer must have
the Credential Provider installed to the Linux host machine where the
integration application is running.

Additional requirements to enable the Credential Provider feature:

* Install Credential Provider on Linux host
  * CP Service
  * CP SDK

* Privilege Cloud Account
  * “CyberArk GitGuardian Integration Host” Platform Installed – Download from the CyberArk marketplace

## Linux Host Configuration

### Linux Host Requirements

* Git and Go lang installed

* Network configured to allow incoming GitGuardian webhook requests

* Front-end proxy installed, NGINX, to proxy pass to the integration and
  termination point for SSL traffic
  * Hostname resolves in DNS
  * SSL Certs created for the endpoint

### Setup The Frontend Proxy

1. Login to a linux host where you will run the integration service

2. Install NGINX

3. Configure NGINX to be a frontend proxy; here is an example
   configuration where the hostname has been added to DNS as
   “webhook1.example.com” and the associated SSL certificates have been
   created and are available on the linux host for NGINX. Example
   config:

  ```nginx
  server {
    listen 443 ssl;
    server_name webhook1.example.com;
    ssl_certificate ssl/webhook1.example.com/webhook1.example.com.cer;
    ssl_certificate_key ssl/webhook1.example.com/webhook1.example.com.key;
    ssl_protocols TLSv1 TLSv1.1 TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    location / {
      proxy_pass http://localhost:9191;
      include proxy_params;
    }
  }
  ```

## GitGuardian Configuration

### GG Requirements

* Endpoint defined for the integration, ex:
  <https://webhook1.example.com/>

### Generate Personal Access Token

This is used when calling HMSL to avoid the daily limit of 5 requests
per day; GitGuardian login url: <https://dashboard.gitguardian.com/>

1) Go to the GitGuardian dashboard and navigate to the API screen
   <img src="./images/integration-guide/image1.png" style="width:4.79444in;height:4.49952in"
   alt="A screenshot of a computer Description automatically generated" />

2) Click on the Create Token button to create the PAT  
   <img src="./images/integration-guide/image2.png" style="width:5.97808in;height:1.08257in"
   alt="A white background with a blue stripe Description automatically generated with medium confidence" />

3) Check the “Scan” checkbox, and click create token  
   <img src="./images/integration-guide/image3.png" style="width:3.84259in;height:6.83338in"
   alt="A screenshot of a computer Description automatically generated" />

   * Copy the token and save it for configuring the integration

## Create the GitGuardian Webhook

This is used when calling HMSL to avoid the daily limit of 5 requests
per day; GG login url: <https://dashboard.gitguardian.com/>

1. Navigate to the Integrations and then to the Secrets Manager
    section  
    <img src="./images/integration-guide/image4.png" style="width:4.34259in;height:2.23718in"
    alt="A screenshot of a computer Description automatically generated" />

2. Click the install button and configure the integration – **Copy the
    webhook token** this will be used in the integration
    configuration.  
    <img src="./images/integration-guide/image5.png" style="width:6.5in;height:4.62151in"
    alt="A screenshot of a webpage Description automatically generated" />

## Install and Build the Integration Service

1. Create a directory to hold the integration code and change into the
    directory

  ```shell
  mkdir work
  cd work
  ```

2. Clone the git repo to the machine

  ```shell
  git clone
  https://github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service.git
  ```

3. Change into the repo directory

  ```shell
  cd cyberark-gitguardian-hmsl-remediation-integration-service
  ```

4. Compile the binaries using make:

  ```shell
  make build-all-bins
  ```

### Configure The Integration Service

1. Obtain connection information for the PostgreSQL server

  ```text
  Example:
    db user: DBUSER
    db pass: DBPASS
    db host: localhost
    db port: 26527
    db name: brimstone (database name)

  Using these parameters would create the connection string:
    -dburl "postgresql://DBUSER:DBPASS@localhost:26257/brimstone"
  ```

2. Obtain connection information for the Privilege Cloud account

  ```text
  Example:
    ID Tenant ID: ID_TENANTID
    Privilege Cloud Tenant ID: PCLOUD_TENANTID
    Privilege Cloud Service User: PCLOUD_SERVICE_USER
    Privilege Cloud Service User Pass: PCLOUD_SERVICE_USER_PASS
    Pending Safe name: PENDING_SAFENAME
    
  Using these parameters to populate the service configuration
  configuration
    -idtenanturl "https://ID_TENANTID.id.cyberark.cloud"
    -pcloudurl "https://PCLOUD_TENANTID.privilegecloud.cyberark.cloud"
    -pamuser "PCLOUD_SERVICE_USER"
    -pampass "PCLOUD_SERVICE_USER_PASS"
    -safename "PENDING_SAFENAME"
  ```

3. Determine the port that the service will listen on; this will be the
port that the NGINX server is configured to proxy pass

  ```text
  Example:
    listen on port 9191

  Using this to populate the configuration parameter
    -port "9191"
  ```

4. Generate an API key that the service will use to authenticate
    incoming requests with

  ```text
  Example:
    generate a dev API key with value: dev123  
    
  Using this to populate the configuration environment variable
    export BRIMSTONE_API_KEY="dev123"
  ```

### Create a startup script

1. Using the configuration information gathered, fill in the
   appropriate fields in the example script, save the script in the
   “./bin” directory

  ```shell
  #!/usr/bin/bash

  # set the value of the API key for authn to brimstone
  export BRIMSTONE_API_KEY="dev123"

  # set this to the value of the GitGuardian API Token
  export GG_API_TOKEN="abc123"

  # set this to the value of the GitGuardian Webhook token presented
  # when configuring the webhook
  export GG_WEBHOOK_TOKEN="xxx"

  ./bin/brimstone -d \
  -hmslurl "https://api.hasmysecretleaked.com" \
  -hmslaudtype "hmsl" \
  -ggapiurl "https://api.gitguardian.com" \
  -ggapitokenvar "GG_API_TOKEN" \
  -ggwebhooktokenvar="GG_WEBHOOK_TOKEN" \
  -keyvar "BRIMSTONE_API_KEY" \
  -dburl "postgresql://DBUSER:DBPASS@localhost:26257/brimstone" \
  -port "9191" \
  -idtenanturl "https://ID_TENANTID.id.cyberark.cloud" \
  -pcloudurl "https://PCLOUD_TENANTID.privilegecloud.cyberark.cloud" \
  -pamuser "PCLOUD_SERVICE_USER" \
  -pampass "PCLOUD_SERVICE_USER_PASS" \
  -safename "PENDING_SAFENAME" \> ./bin/brimstone.log 2\>
  ./bin/brimstone-err.log
  ```

2. Run the script to check that everything is configured properly

  ```shell
  bash ./bin/local-start-script.sh
  ```

3. Check the logs to verify it is running properly

  ```shell
  tail ./bin/brimstone.log
  # Check log output

  tail ./bin/brimstone-err.log
  # Check log output
  ```

## Credential Provider Installation

Install the Credential Provider on the same linux host where you will
run the integration application. Follow this link and follow the
instructions to install the CP:

[Credential Provider
Documentation](https://docs.cyberark.com/credential-providers/latest/en/Content/CP+and+ASCP/lp_cp.htm)

URL:
<https://docs.cyberark.com/credential-providers/latest/en/Content/CP+and+ASCP/lp_cp.htm>

### Credential Provider Deamon Installation

Installation of CP service:

<https://docs.cyberark.com/credential-providers/latest/en/Content/CP+and+ASCP/Installation-on-AIX-RHEL-CentOS.htm>

### Credential Provider SDK Installation

Installation of SDK:

<https://docs.cyberark.com/credential-providers/latest/en/Content/CP%20and%20ASCP/Working-with-Application-Password-SDK.htm>

Configuring the SDK:

Note: where it says to copy the files into a `/usr/lib` directory, copy
the files, and do not symlink to the files.

<https://docs.cyberark.com/credential-providers/latest/en/Content/CP%20and%20ASCP/C-Application-Password-SDK.htm?tocpath=Developer%7CCredential%20Provider%7CApplication%20Password%20SDKs%7CC%252FC%252B%252B%20Application%20Password%20SDK%7C_____0#SetuptheCCApplicationPasswordSDK>

## Install and Build the Integration Service with Credential Provider

1. Create a directory to hold the integration code and change into the
   directory

  ```shell
  mkdir work
  cd work
  ```

2. Clone the git repo to the machine

  ```shell
  git clone
  https://github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service.git
  ```

3. Change into the repo directory

  ```shell
  cd cyberark-gitguardian-hmsl-remediation-integration-service
  ```

4. Compile the binaries using make:

  ```shell
  make build-all-bins
  ```

### Install the CyberArk GitGuardian Integration Host Platform

* Download the zip file from the CyberArk marketplace

* Import into the Privilege Cloud account

  Go to: Administration -> Platform Management -> Import platform

  <img src="./images/integration-guide/image6.png" style="width:2.17593in;height:3.58147in"
    alt="A screenshot of a computer Description automatically generated" />

* The new platform should show up under “Applications”

  <img src="./images/integration-guide/image7.png" style="width:6.5in;height:1.69444in"
alt="A screenshot of a computer Description automatically generated" />

## Configure The CP Integration Service

* Create account for database using PostgreSQL platform

  Go To Accounts View

  <img src="./images/integration-guide/image8.png" style="width:2.02778in;height:4.2125in"
  alt="A screenshot of a phone Description automatically generated" />

  Click on Add Account

  <img src="./images/integration-guide/image9.png" style="width:3.82407in;height:0.64838in"
  alt="A white background with black and white clouds Description automatically generated" />

  1. Select system type “Database”

     <img src="./images/integration-guide/image10.png" style="width:1.9537in;height:0.87384in"
     alt="A screenshot of a computer Description automatically generated" />

  2. Select platform “PostgreSQL”

     <img src="./images/integration-guide/image11.png" style="width:1.95347in;height:0.88472in"
     alt="A screenshot of a computer Description automatically generated" />

  3. Select safe where to add the account, for this example, we have
     created safename “IntegrationHostSafe” – This is the safe where
     the integration credentials will be stored

     <img src="./images/integration-guide/image12.png" style="width:2.06482in;height:0.96459in"
     alt="A screenshot of a screen Description automatically generated" />

  4. Fill in the properties for the database, use the values gathered
     from setting up the database step, then click save button

     <img src="./images/integration-guide/image13.png" style="width:2.71296in;height:4.93455in"
     alt="A screenshot of a computer screen Description automatically generated" />

* Create account for integration service using CyberArk GitGuardian
  Integration Host platform

  1. Select system type “Application”

     <img src="./images/integration-guide/image14.png" style="width:2.12963in;height:0.66728in"
     alt="A screenshot of a computer application Description automatically generated" />

  2. Select platform “CyberArk GitGuardian Integration Host”

     <img src="./images/integration-guide/image15.png" style="width:2.39815in;height:0.77466in"
     alt="A close-up of a sign Description automatically generated" />

  3. Select safe where to add the account, for this example, we have
     created safename “ExamplePendingSafe” – This is the safe where the
     integration credentials will be stored

     <img src="./images/integration-guide/image12.png" style="width:2.06482in;height:0.96459in"
     alt="A screenshot of a screen Description automatically generated" />

  4. Fill in the properties for the integration host, use the values
     gathered from setting up the linux host step, then click the save
     button

     <img src="./images/integration-guide/image16.png" style="width:3.09039in;height:8.14815in"
     alt="A screenshot of a computer Description automatically generated" />

### Create a Privilege Cloud Application

1) Navigate to Applications

   <img src="./images/integration-guide/image17.png" style="width:1.90741in;height:2.83498in"
   alt="A screenshot of a phone Description automatically generated" />

2) Click on Add Application

   <img src="./images/integration-guide/image18.png" style="width:3.72222in;height:1.15206in"
   alt="A screenshot of a computer Description automatically generated" />

3) Add a name for the application, this will be used in the
   configuration in the startup script, placeholder in the startup
   script is INTEGRATION-HOST-APPLICATION-NAME, in the screenshot the
   value from the “Name:” field is what you will use in the startup
   script.

   <img src="./images/integration-guide/image19.png" style="width:3.99319in;height:3.49074in"
   alt="A screenshot of a application Description automatically generated" />

## Create a CP startup script

1. Using the configuration information gathered, fill in the
   appropriate fields in the example script, save the script in the
   “./bin” directory

  ```shell
  #!/usr/bin/bash

  ./bin/brimstone-cp -d \
  -safename "SAFENAME-WHERE-INTEGRATION-HOST-SETTINGS-ARE-STORED" \
  -hostobjname "CYBR-GG-INTEGRATION-HOST-ACCOUNT-NAME" \
  -dbobjname "POSTGRESQL-ACCOUNT-NAME" \
  -appid "INTEGRATION-HOST-APPLICATION-NAME" \</dev/null \>
  ./bin/brimstone-cp.log 2\> ./bin/brimstone-cp-err.log
  ```

2. Run the script to check that everything is configured properly

  ```shell
  bash ./bin/local-start-script-cp.sh
  ```

3. Check the logs to verify it is running properly

  ```shell
  tail ./bin/brimstone-cp.log
  # Check log output

  tail ./bin/brimstone-err-cp.log
  # Check log output
  ```
