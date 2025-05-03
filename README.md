# Dr. Task
A simple task management app that can utilise a local AI

## Development
### Logging
To enable debug and message logs you need to set the `DEBUG` environment variable.
On MacOS use
```terminal
export DEBUG=1
```

Log outputs are sent to `debug.log`.

It can be useful to see all `tea.Msg`'s sent to the main models `Update()` function.
They are automatically sent to `messages.log`. To keep track, in another terminal you can run
```terminal
tail -f messages.log
```
