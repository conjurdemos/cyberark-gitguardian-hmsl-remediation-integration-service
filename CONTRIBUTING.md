<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

* [Contributing to the Brimstone Repo](#contributing-to-the-brimstone-repo)
  * [Table of Contents](#table-of-contents)
  * [Where Do I Start?](#where-do-i-start)
  * [Contribution Workflow](#contribution-workflow)
    * [Reporting an Issue](#reporting-an-issue)
    * [Finding Issues to Work On](#finding-issues-to-work-on)
    * [Reviewing license requirements](#reviewing-license-requirements)
      * [When the Repo Includes the CLA](#when-the-repo-includes-the-cla)
      * [When the Repo Does Not Include the CLA](#when-the-repo-does-not-include-the-cla)
        * [Signing Off a Commit](#signing-off-a-commit)
        * [Verifying That You Have Signed a Commit](#verifying-that-you-have-signed-a-commit)
    * [Working on Issues](#working-on-issues)
    * [Submitting a Pull Request](#submitting-a-pull-request)
    * [Getting Your Code Reviewed](#getting-your-code-reviewed)
  * [Contributing to this project](#contributing-to-this-project)
  * [Tips For Making a Project Open to Contributions](#tips-for-making-a-project-open-to-contributions)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Contributing to the Brimstone Repo

Whether you're a member of the CyberArk team, or are just getting involved with our community, these
general-purpose guidelines are meant to help you familiarize yourself with our processes. You can
use them when working on any of our public projects, including this one!

## Table of Contents


## Where Do I Start?

As you visit the various guides and documents throughout this repository, you may notice something
that you feel is missing, or perhaps you feel it could be improved in some way. This documentation
is used by a wide variety of community members, it's important to us that we are as clear and
thorough as possible, so your perspective is important! Your feedback can help countless others, so
don't be afraid to be the change you want to see!

## Contribution Workflow
Your first step should be checking the `Issues` tab to see if someone else has already created the
issue. If it has been created, check out the [working on issues](#working-on-issues) section, or
feel free to add additional context or information about the topic. If it hasn't been created, read
about [how to report a new issue](#reporting-an-issue)!


### Reporting an Issue
1. After confirming that the topic doesn't already exist, you'll need to click `New Issue` to create
   it.

1. Select a template, as this will automatically add an appropriate tag and provide additional
   guidelines for what information is required.

1. Create a descriptive title for the issue, using present-tense verbage, e.g.
    > Secretless Broker does not have emoji support

1. Add the appropriate tag for your issue, if one was not added by the template, e.g.
    > kind/documenation kind/enhancement

1. Continue to monitor the issue for responses and engage in conversation until resolved

### Finding Issues to Work On

1. Click the `Issues` tab on Github to see open issues

1. Select an existing issue and provide the following information in a comment:
    - State your intention to work on the issue
    - Provide an outline for your implementation plan  

    This is important to avoid the overlapping of work!

1. If the issue is already in progress, feel free to ask questions or engage in conversation about
    the implementation. Otherwise, review the repository license requirements
    before starting work and continue on to [working on issues](#working-on-issues).

### Reviewing license requirements

To contribute to any CyberArk open source project, you must make sure you've met
our contributor license requirements.

To find out what the project you're interested in requires, refer to the
`CONTRIBUTING.md` file in the repository.

#### When the Repo Includes the CLA

If the `CONTRIBUTING.md` document indicates that you need to sign the CLA, for example:

> "Thanks for your interest in Conjur. Before contributing, please take a moment to read and sign
> our Contributor Agreement. This provides patent protection for all Conjur users and allows
> CyberArk to enforce its license terms. Please email a signed copy to <oss@cyberark.com>."

Then you **must** submit a signed copy of the CLA to <oss@cyberark.com> in order
for your contribution to be merged.

A copy of the CLA can be found [here](https://github.com/cyberark/community/blob/main/documents/CyberArk_Open_Source_Contributor_Agreement.pdf).

#### When the Repo Does Not Include the CLA

If the repo does not ask you to sign the CLA, then you may do one of the following:

-  Follow the above instructions for [signing the CLA](#when-the-repo-includes-the-cla).

**OR**

-  Sign your commits using the `--signoff` flag in the `git` CLI (more info
   [here](https://git-scm.com/docs/git-commit#Documentation/git-commit.txt--s)) to certify that you
   have the rights to submit the work under the same license as the project and you agree to a
   Developer Certificate of Origin (see http://developercertificate.org/ for more information).

##### Signing Off a Commit

When signing, every commit must be signed, and your PR will not be approved until this is the case.

To sign-off, use:
```
git commit --signoff
```
or
```
git commit -s
```

Alternatively, if you have already made the commit, you may use:
```
git commit --amend --signoff
```
or
```
git commit --amend -s
```

After amending a commit, you will need to force-push (`git push -f <branch_name>`)
to your branch for this to take effect.

> Note: Only repositories licensed under MIT, BSD, or Apache 2.0 are eligible to
  use the DCO instead of the CLA for community contributions!! **If the repo
  does not meet this criterion, you _must_ sign the CLA.**

##### Verifying That You Have Signed a Commit

To determine if you **have** signed a commit, look at the commit history in the branch to ensure
each commit includes a line like:

    Signed-off-by: Jane A Doe <jane@developer.example.org>

This line should include your real name and contact email.

### Working on Issues

1. After receiving confirmation that the issue is not already in progress, add the `Implementing`
   label to the issue as you begin to work on it.

1. [Fork the project](https://help.github.com/en/github/getting-started-with-github/fork-a-repo).

1. [Clone your fork](https://help.github.com/en/github/creating-cloning-and-archiving-repositories/cloning-a-repository).

1. Make local changes to your fork by editing files.

1. Run tests as described in the `CONTRIBUTING.md` or as outlined in the test stage(s) in the
   `Jenkinsfile`, ensuring they pass.

1. [Commit your changes](https://help.github.com/en/github/managing-files-in-a-repository/adding-a-file-to-a-repository-using-the-command-line), making sure to
   [sign off on your commits](#signing-off-a-commit).

### Submitting a Pull Request

1. Once you have wrapped up your work, you should verify that your commit history is
   [clean and organized](https://www.conjur.org/blog/cleaning-history-for-github-prs/).

1. [Sync your fork with the upstream repository](https://gist.github.com/ravibhure/a7e0918ff4937c9ea1c456698dcd58aa)
   and resolve any conflicts.

1. [Push your local changes to the remote server](https://help.github.com/en/github/using-git/pushing-commits-to-a-remote-repository).

1. Navigate to GitHub and [open a `Pull Request`](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork) ("`PR`").

1. Make sure you're requesting a merge into the main (or default) branch of the upstream project from your fork.

1. Add a PR title and description.
    + Add a title to the PR that describes what the set of changes included in the PR accomplishes,
      as well as a reference to the issue it addresses, if an issue exists. The PR title should be
      clear, specific, and succinct.

        * Examples of bad PR titles:

                Component refactor
                Minor fixes
                Add tests

        * Examples of good PR titles:

                #40 - Improved logs for default authenticator
                Bump special-gem version to 1.3.2
                #1645 - Add support for configuring connector plugin via environment

    + Add a description to the PR that details the changes you made, filling out any categories
      listed by a template, and linking the original issue in the description, e.g.
        > Connected to #[Issue Number]

1. Add the `Implemented` label to the issue.

1. Tag a reviewer if any are recommended in the righthand menu.

1. Submit the Pull-Request.

### Getting Your Code Reviewed
Once your pull-request has been submitted for review, a reviewer will look for a few things.

- A maintainer will verify that any contributor license agreements have been met. Required
  contributor license agreements (if any) are usually documented in the repo's CONTRIBUTING.md.

- A maintainer will review your code, and verify that it addresses the issue sufficiently while
  meeting the contribution guidelines or style patterns for the project.

  Here are a few things a reviewer may look for.
   - Has the branch author done a good job cleaning up the commit history?
   - Has the branch been rebased against the main or default branch?
   - Has the code been thoroughly tested?
      - Have you ran any unit or integration tests, and made note of it in the pull-request
        description?
      - Have you performed any manual or non-standard tests, and made note of them in the
        pull-request description?
   - Does this pull-request require an update to any existing documentation?
      - If so, does an issue need to be made for updating the documentation?
   - Is the code written in a way that is consistent with the existing code base?

- The maintainer will finalize their review
   - If a reviewer notices an issue with your pull-request, such as those listed above, they will
     request changes be made. You may need to go back and fix a thing or two before it can be
     approved.
   - If any of the reviewer's feedback is optional, it will be clearly marked "Not required" or
     "nit". It might still be worth addressing, if you can.

- Respond directly to any comments with any of your questions or concerns.

- After making any changes, required or otherwise, make sure your commit history is still organized,
  and make sure you have recently rebased off of main or default branch.

- If changes were requested, you can prompt for a re-review at the bottom of your pull-request's
  page in Github. Otherwise, you can click the 're-review' icon beside the reviewers name on the
  righthand side of your pull-request page.

- From here you will engage in conversation with the reviewer(s) until they are satisfied with the
  branch state, and confirm that it is ready to merge into main or default branch. Once the PR is approved (Marked
  as "Approved" in Github), you are ready to merge your changes!

## Contributing to this project
We want to provide our community with the best possible experience when they choose to contribute to
one of our projects. That means constant improvements to this project and the documentation therein.

Looking through [issues for this project](https://github.com/conjurdemos/cyberark-gitguardian-hmsl-remediation-integration-service/issues) is a great
way to see what we're doing to improve the community experience, and to see what we're planning for
the future.

If you have an idea for something that could be added or improved, feel free to submit an issue
yourself!

## Tips For Making a Project Open to Contributions
Since standards for contribution can vary from project to project, it's beneficial to use the
following tips to make that information easy to find, and easy to use.

1. The project should be listed in the dropdown-list in [the Brimstone README.md](README.md), under
   its respective team

1. Within the project's own repository, a CONTRIBUTING.md file should be the standard entry-point
   for a contributor, whether internal or external. If it does not exists, it's recommended that one
   is created, and pointed to by the top-level README.md of the project. </br> A good
   CONTRIBUTING.md has the following:
    - A reference to this community repo for general guidelines
    - Where to find issues
    - How to report new issues
    - How to work on issues
    - Standards for writing in the project-specific language
    - How to build, run, and test the project
    - Any relevant forums for discussion and collaboration
    - Any project-specific best practices

1. Templates for issues can improve the information gathered for a new issue, and provide helpful
   organization, while reducing the amount of domain knowledge required for filing an issue, and
   should be created and wherever possible. For example, templates for common cases like `bugs` or
   `feature requests` can save time and cut down on requests for additional information.

