# How to run?

```bash
# start docker-compose
docker compose up -d
```

```bash
# Run the mock-client for testing
go run ./client/cmd
```

## How to test?

```bash
cd ./notification/
make test/cover # runs all unit-tests and opens a coverage view on your default browser
make integration-test/cover # runs all unit-tests and opens a coverage view on your default browser
```

## The Challenge

### Backend Rate-Limited Notification Service

We have a Notification system that sends out email notifications of various types
(status update, daily news, project invitations, etc). We need to protect
recipients from getting too many emails, either due to system errors or due to
abuse, so let’s limit the number of emails sent to them by implementing a
rate-limited version of **_NotificationService_**.

The system must reject requests that are over the limit.

Some sample **notification types** and [rate limit rules], e.g.:

- **Status**: [not more than 2 per minute for each recipient]

- **News**: [not more than 1 per day for each recipient]

- **Marketing**: [not more than 3 per hour for each recipient]

- Etc. these are just samples, the system might have several rate limit rules!

> With this in mind, I suppose that they want us to be able to add easily,
> either in Code and/or configuration-files possible rules and types

```java
import java.io.*;

import java.util.*;

class Solution {
  public static void main(String[] args) {
    NotificationServiceImpl service = new NotificationServiceImpl(new Gateway());
    service.send("news", "user", "news 1");
    service.send("news", "user", "news 2");
    service.send("news", "user", "news 3");
    service.send("news", "another user", "news 1");
    service.send("update", "user", "update 1");
  }
}

interface NotificationService {
  void send(String type, String userId, String message);
}

class NotificationServiceImpl implements NotificationService {
  private Gateway gateway;
  public NotificationServiceImpl(Gateway gateway) {
  this.gateway = gateway;
}

// TASK: IMPLEMENT this
@Override
public void send(String type, String userId, String message) {
  throw new RuntimeException("not implemented - fix this");
}

class Gateway {

  void send(String userId, String message) {
    System.out.println("sending message to user " + userId);
  }
}
```
