# Toy Cloud

This project is a toy implementation of some core cloud and
distributed computing services. Each folder contains a single
service that can be a part of the overall architecture.

- [] Web Service (Dummy Service)
- [] Load Balancer
- [] Manual Scaler
- [] Auto Scaler
- [] Service Discovery
- [] Dashboard

Servers in the "cloud" are simulated with docker containers, where
1 docker container represents 1 server running 1 service.
As such, it's necessary for the containers to be networked with
each other using a user-defined bridge. The docker-compose.yaml
should help you spin everything up already connected to each
other.