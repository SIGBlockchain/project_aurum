<img src = "media/aurum_logo_readme.jpg"  alt="drawing" width="200"/>

Project Aurum
=============
Aurum is a proprietary UIC ACM blockchain project. The current planned use case is a token to be exchanged among students.

## Motivation
To create a working blockchain and learn about the technology behind it. SIG Blockchain Developers use software engineering techniques to construct a complex, collaborative, and efficient product.

## Tech/framework used

<b>Built with</b>
- [Golang](https://golang.org/)
- [Docker](https://www.docker.com/)

## Features
Aurum Phase One is a centralized private blockchain that provides users with the ability to exchange tokens with each other. Future support for decentralizing Aurum will be implemented once the network grows to a sustainable level for decentralization.

## Installation
`under development`

## API Reference
`under development`

## Tests
Run `go test -v` at project root.

## Docker
To build an image of the producer:
    -change current directory to project_aurum/producer
    -run `docker build -t <name:tag> .`
To run an instance of the producer: 
    -`docker run -p 13131:13131 <what ever you named the image>`

To run the compose file:
    -`docker-compose run --service-ports producer`

## How to use?
`under development`

## Contribute
If you would like to contribute, please comment on an issue you'd like to take on. Then, make a branch based on `dev`. Once you've completed the issue make a pull request from your branch to `dev`. If you have any questions simply ask in a comment on the issue.

## Credits
- First and foremost a big thank you to everyone who has worked on the SIG Blockchain team, both past and present. 
- Thank you to the [Association of Computing Machinery UIC Chapter](https://acm.cs.uic.edu/), for providing the support that SIG Blockchain has needed to thrive.
- Shoutout to [calvinmorett](https://github.com/calvinmorett) for providing us with the awesome Aurum logo.

## License
MIT License
- See the [LICENSE](https://github.com/SIGBlockchain/project_aurum/blob/readme/LICENSE) file for details.
