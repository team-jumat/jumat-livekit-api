##
## STEP 1 - BUILD
##

# specify the base image to  be used for the application, alpine or ubuntu
FROM golang:1.20.0-alpine AS build

RUN apk --no-cache add tzdata

# create a working directory inside the image
WORKDIR /livekit-api/livekit-api

# copy all to image
COPY . /livekit-api

# download Go modules and dependencies
RUN go mod download

RUN ls

# compile application
RUN go build -o /app

##
## STEP 2 - DEPLOY
##

FROM scratch

WORKDIR /

COPY --from=build /livekit-api/database/.env /database/.env

COPY --from=build /livekit-api/livekit-api/.env /.env

COPY --from=build /livekit-api/log /log

COPY --from=build /app /app

CMD ls

# tells Docker that the container listens on specified network ports at runtime
EXPOSE 8080

# command to be used to execute when the image is used to start a container
CMD ["/app"]

