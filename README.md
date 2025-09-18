# Backend Rate-Limited Notification Service

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

NOTES:
Your solution will be evaluated on code quality, clarity and development best practices.

Take into account that this is the initial assessment of the work quality that
you can produce in a productive environment so show your best face.

Feel free to use the programming language, frameworks, technologies, etc that
you feel more comfortable with.

Below you'll find a code snippet that can serve as a guidance of one of the
implementation alternatives in Java. Feel free to use it if you find it useful
or ignore it otherwise; it is not required to use it at all nor translate this
code to your programming language.

Please share with us the link of the repository containing the application source
code when finished and let us know to schedule the meeting to review the app.

Feel free to ping me any question you might have about this technical assessment.

Code snippet sample

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
