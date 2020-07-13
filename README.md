# TweetWrap

TweetWrap handles the 3-legged OAuth process and auth state in order to simplify the initialization of [dghubble's go-twitter](https://github.com/dghubble/go-twitter) library.

## How to use

### Regiter your app

First you need a valid [Twitter developer account](https://developer.twitter.com/en/apply). Create your project and get your `API key` and `API secret key`.

### First run

Given the following code:

```golang
twitt, authURL, err := tweetwrap.New(tweetwrap.Config{
    APIKey:       "yours here",
    APIKeySecret: "yours here",
})
if err != nil {
    log.Fatalln(err)
}
if authURL != nil {
    fmt.Println("Authentication proceedure initiated, please logging at:", authURL.String())
    fmt.Println("Then restart with the config updated with the validation PIN.")
    os.Exit(0)
}
fmt.Println("Authenticated as:", twitt.GetAuthedUser())
```

As no current state has been found by the wrapper, it starts a 3-legged OAuth process with your app identity. Following the URL and validating the OAuth consent form will allow your app to act as the user you were logged as. The consent form will yield a PIN verification code which you will need on the second run.

### Second run

```golang
twitt, authURL, err := tweetwrap.New(tweetwrap.Config{
    APIKey:       "yours here",
    APIKeySecret: "yours here",
    PIN:          "yours here", // <--- NEW
})
if err != nil {
    log.Fatalln(err)
}
if authURL != nil {
    fmt.Println("Authentication proceedure initiated, please logging at:", authURL.String())
    fmt.Println("Then restart with the config updated with the validation PIN.")
    os.Exit(0)
}
fmt.Println("Authenticated as:", twitt.GetAuthedUser())

// client is now initialized by the wrapper and can be used directly
tweet, _, err := twitt.Client.Statuses.Update("Yay wrapped tweet !", nil)
// ...
```

On the second run, the wrapper will load the previous state and detect the ongoing authentification. It will use the PIN code in order to finalize the auth between the app and the user. If authentification went fine, the linked user name will be printed. And the [dghubble's go-twitter](https://github.com/dghubble/go-twitter) client accessed directly with the `Client` wrapper member.

### Customizing the state file location

By default state will be saved within `token_credentials.json` file on the current working directory. Path can be customized by submitting a non empty `StateFilePath` value within `Config`.
