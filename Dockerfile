FROM golang:1.19.2


COPY . /app
WORKDIR /app
RUN make setup
RUN make generate
#ENV PORT 8080
#EXPOSE $PORT
USER root
CMD ["go" ,"run" ,"./cmd/swagger/"]
