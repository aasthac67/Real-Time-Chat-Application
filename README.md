# Real-Time Chat Application

## Overview

Real-Time Chat Application is a robust chat application built with React for the frontend and Go for the backend, utilizing PostgreSQL for database storage and Redis for caching. The application is containerized using Docker and orchestrated with Kubernetes. It supports real-time communication and features upvote/downvote functionality using WebSockets.

## Features

- **User Authentication:**

  - Users can sign up and log in.
  - New usernames and passwords are stored in the database with passwords encrypted using bcrypt.
  - Authentication ensures correct username and password entry, with additional checks for passwords being between 8 to 20 characters during signup.

- **User List:**

  - Once logged in, users can view a real-time updated list of all other registered users.
  - New users appearing in the system are instantly reflected in the user list of any other logged-in users.

- **Chat Functionality:**
  - Users can select any other user to start a chat.
  - Messages are sent and received in real time.
  - Upvotes and downvotes on messages are also updated in real time.
  - Users can see chat history as well.
- **Upvote and Downvote:**

  - Each user can upvote or downvote messages.
  - A user can only cast one vote (upvote or downvote) per message.
  - Selecting an already chosen vote (upvote or downvote) removes it. Switching between upvote and downvote is seamlessly handled.
  - Horizontal scaling of upvote/downvote functionality is managed using Redis as a cache to store message IDs with their votes, minimizing frequent database calls.

- **Concurrency and Data Integrity:**

  - Race conditions for upvotes and downvotes are managed using WebSockets.
  - Votes are stored in a separate database, with asynchronous functions ensuring consistent and reliable vote counts.

- **Data Privacy and Security:**
  - Messages, upvotes, and downvotes are strictly confined to the intended users.
  - No user can access messages, upvotes, or downvotes that were not meant for them.

This feature set ensures a smooth and secure user experience with real-time interaction and robust handling of intensive functions like upvoting and downvoting.

## Tech Stack

- **Frontend:** React Typescript
- **Backend:** Go
- **Database:** Postgres
- **Protocol:** RESTful API, Websockets
- **Message Queue:** Redis
- **Local Deployment:** Docker
- **Infrastructure:** Minikube

## Setup to run locally

1. Runs docker containers for Backend, Frontend, Postgres and Redis

`docker compose up`

The frontend can be accessed at http://localhost:3000/ after the above command is run.

## Kubernetes Deployment

1. Start Minikube

`minikube start`

2. Set Docker to use Minikube's Docker daemon

`minikube docker-env`

`eval $(minikube -p minikube docker-env)`

3. Build and pull all docker images

`docker compose build`

`docker compose pull`

4. Deploy Backend, Frontend, Postgres and Redis into the minikube cluster

`kubectl apply -f deployments`
