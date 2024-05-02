<img alt="CyberArk Banner" src="images/cyberark-banner.jpg">

<!--
Author:   David Hisel <david.hisel@cyberark.com>
Updated:  <2024-02-21 16:38:42 david.hisel>
-->

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

* [Brimstone Privilege Cloud Notes](#brimstone-privilege-cloud-notes)
  * [Add Safe](#add-safe)
  * [Create Brimstone Service User](#create-brimstone-service-user)
  * [Associate Service User To Safe](#associate-service-user-to-safe)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Brimstone Privilege Cloud Notes

## Add Safe

In the Privilege cloud UI, navigate to Policies -> Safes, and then click the "Create Safe" button.

1. Assign safe name
2. Add member(s) -- for brimstone create a user that has full permissions
3. Click the Create safe button

## Create Brimstone Service User

In the Identity Administration UI, navigate to Core Services -> Users, and then click the "Add User" button.

1. Enter login name, Display name, and password (required fields with a "*")
2. Check the Password never expires checkbox
3. Check the Is service user checkbox
4. Check the Is Oauth confidential client checkbox
5. Click the add user button

## Associate Service User To Safe

Note: human user credentials will not work because the API limits access to service users.

In the Privilege cloud UI, navigate to Policies -> Safes -> search for your safe.

1. Click on the line that contains your safe to highlight it, and bring up the sub-pane, this should show 2 tabs, the "Details" and the "Members" tabs.
2. Click on the "Members" tab
3. Click on the "Add members" button
4. Search for your user to add, then select the user by checking the checkbox next to the user
5. After selecting your user, click on the next button
6. Click on "Full" in The "select bar" at the top in the Permissions Presets.
7. Click on the Add button
