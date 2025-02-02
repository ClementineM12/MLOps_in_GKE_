# Helm Charts

## Tools Used
Since helm chart are a challenge both in navigating the code and validating their rendered manifests, we are using a set of tools to make the process friendlier and somehow easier.

### Documentation
Documentation for our charts is generated using the [helm-docs](https://github.com/norwoodj/helm-docs) tool. The tool is using doc strings inside the values file of the chart to generate a table of supported values along with their types and defaults.

### Install the tool
To install the tool run any of the command below
```sh
# Using brew (recommended)
brew install norwoodj/tap/helm-docs
# Using golang
go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
```

### Generate the docs
To generate the readme file for a chart just run.
```sh
helm-docs -o values.md
```

### Tests
Tests are based on the [unittest](https://github.com/helm-unittest/helm-unittest) framework. The framework provides a yaml based interface to perform unit like tests in helm charts. More specifically it offers a way to provide a set of values (on top of the default `values.yaml` file of the chart ) and then describe what's expected to be rendered. The full list of configuration options and test cases can be found [here](https://github.com/helm-unittest/helm-unittest/blob/main/DOCUMENT.md).

### Install the tool
The tool can be installed as a `helm` plugin by running the command below
```sh
helm plugin install https://github.com/helm-unittest/helm-unittest.git
```

### Run the tool
To run the test suites of chart locally, after navigating to the chart's folder run the following command
```sh
helm unittest .
```

## Changelog Generation
To keep track of changes between Helm Chart versions, we use the [helm-changelog](https://github.com/mogensen/helm-changelog) tool. This tool helps generate a changelog whenever a version is bumped, ensuring that all modifications, enhancements, and fixes are documented in a consistent manner. To use this tool, install it by running (Go is required):
```sh
go install github.com/mogensen/helm-changelog@latest
```

After installing the tool, you should be able to run it if your $GOPATH/bin is in your PATH. If not, add it as follows:
```sh
export PATH=$PATH:$(go env GOPATH)/bin
```

After bumping the version of a chart, you can generate the changelog by executing inside the folder:
```sh
helm-changelog
```

This will create or update the CHANGELOG.md file with the details of the changes made in the latest version.