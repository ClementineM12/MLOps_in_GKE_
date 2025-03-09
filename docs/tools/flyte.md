# Flyte Guide

## Projects 

Create your first project: 
* Define a yaml file which will store the details of your project and then run:
```sh
    flytectl create project --file <your_project_file>
```
* Get your projects:
```sh
    flytectl get projects
```

## Run workflow
Once registered, you can launch the workflow using the Flyte CLI. For example, to run the workflow with the CNN branch, use:
```sh
flytectl launch --project <project> --domain <domain> --workflow skin_cancer_workflow --inputs '{"model_type": "cnn"}'
```
Similarly, to run the Random Forest branch, either omit the parameter (if "random_forest" is the default) or explicitly set it:
```sh
flytectl launch --project <project> --domain <domain> --workflow skin_cancer_workflow --inputs '{"model_type": "random_forest"}'

```