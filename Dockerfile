FROM golang:1.21
# Set destination for COPY
WORKDIR /app

# Download Go modules
# COPY go.mod go.sum ./
COPY go.mod go.sum /app/

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY . /app/
RUN go mod vendor

# Build
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cmd/ddm/ddm-server.go
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cmd/cli/cli.go

RUN CGO_ENABLED=0 GOOS=linux go build /app/cmd/ddm/ddm-server.go
RUN CGO_ENABLED=0 GOOS=linux go build /app/cmd/cli/cli.go

# sudo docker container create --name ddm ddm
# sudo docker container cp ddm:/app/ddm-server .
# sudo docker container cp ddm:/app/cli .