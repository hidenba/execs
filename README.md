# execs
Interactively select arguments to connect with ECS Exec to connect to Fargate.
# Require

- [session-manager-plugin](https://github.com/aws/session-manager-plugin)

# Install

## Go

```
go install github.com/hidenba/execs@latest
```

# Usage

```
$ execs
```

## Options

### region

Default is ap-northeast-1

```
$ execs -r us-east-2
```

### Profile

```
$ execs -p staging-profile
```

# Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are greatly appreciated.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement". Don't forget to give the project a star! Thanks again!

1. Fork the Project
1. Create your Feature Branch (git checkout -b feature/AmazingFeature)
1. Commit your Changes (git commit -m 'Add some AmazingFeature')
1. Push to the Branch (git push origin feature/AmazingFeature)
1. Open a Pull Request

# License

Distributed under the MIT License. See LICENSE for more information.
