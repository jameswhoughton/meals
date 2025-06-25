# Kubernetes Deployment

**Note This is purely academic, the application is not deployed to prod using Kubernetes.**

This directory contains configuration files to deploy the application to a Kubernetes cluster, the application and database are orchastrated by Kubernetes, the database has a persistant volume to ensure data is retained.

Due to the session cookie parameters, the application will only work over SSL (or localhost).
