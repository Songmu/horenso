# Changelog

## [v0.9.0](https://github.com/Songmu/horenso/compare/v0.3.0...v0.9.0) (2019-02-15)

* introduce go modules [#30](https://github.com/Songmu/horenso/pull/30) ([Songmu](https://github.com/Songmu))
* add --config option [#29](https://github.com/Songmu/horenso/pull/29) ([Songmu](https://github.com/Songmu))
* [incompatible] quit to use pointer for ExitCode. use -1 as default value [#28](https://github.com/Songmu/horenso/pull/28) ([Songmu](https://github.com/Songmu))
* enhance tests around logs [#27](https://github.com/Songmu/horenso/pull/27) ([Songmu](https://github.com/Songmu))
* [incompatible] Reduce pointer from type Report [#25](https://github.com/Songmu/horenso/pull/25) ([Songmu](https://github.com/Songmu))
* introduce Songmu/timestamper [#26](https://github.com/Songmu/horenso/pull/26) ([Songmu](https://github.com/Songmu))

## [v0.3.0](https://github.com/Songmu/horenso/compare/v0.2.0...v0.3.0) (2019-02-10)

* Update deps [#24](https://github.com/Songmu/horenso/pull/24) ([Songmu](https://github.com/Songmu))
* add --log option [#23](https://github.com/Songmu/horenso/pull/23) ([Songmu](https://github.com/Songmu))
* refine timestamp format in timestampWriter [#22](https://github.com/Songmu/horenso/pull/22) ([Songmu](https://github.com/Songmu))
* Add verbose option to output `horenso` itself log output [#21](https://github.com/Songmu/horenso/pull/21) ([Songmu](https://github.com/Songmu))
* refactorings [#20](https://github.com/Songmu/horenso/pull/20) ([Songmu](https://github.com/Songmu))
* update README [#18](https://github.com/Songmu/horenso/pull/18) ([dozen](https://github.com/dozen))

## [v0.2.0](https://github.com/Songmu/horenso/compare/v0.1.0...v0.2.0) (2018-01-16)

* update deps (wrapcommander) [#17](https://github.com/Songmu/horenso/pull/17) ([Songmu](https://github.com/Songmu))
* add override command status option [#16](https://github.com/Songmu/horenso/pull/16) ([dozen](https://github.com/dozen))

## [v0.1.0](https://github.com/Songmu/horenso/compare/v0.0.2...v0.1.0) (2017-12-18)

* add `signaled` field to Report [#15](https://github.com/Songmu/horenso/pull/15) ([Songmu](https://github.com/Songmu))
* adjust releng [#14](https://github.com/Songmu/horenso/pull/14) ([Songmu](https://github.com/Songmu))
* Add Go 1.6, 1.7, and tip to .travis.yml [#13](https://github.com/Songmu/horenso/pull/13) ([achiku](https://github.com/achiku))

## [v0.0.2](https://github.com/Songmu/horenso/compare/v0.0.1...v0.0.2) (2016-03-31)

* cmd.Start() before stdinPipe.Write() [#11](https://github.com/Songmu/horenso/pull/11) ([fujiwara](https://github.com/fujiwara))
* fix typo in Usage [#10](https://github.com/Songmu/horenso/pull/10) ([fujiwara](https://github.com/fujiwara))
* wait to close stdout/err before reporting just to be safe [#9](https://github.com/Songmu/horenso/pull/9) ([Songmu](https://github.com/Songmu))

## [v0.0.1](https://github.com/Songmu/horenso/compare/6c9c2a74...v0.0.1) (2016-01-10)

* Update README [#8](https://github.com/Songmu/horenso/pull/8) ([ariarijp](https://github.com/ariarijp))
* record user/system time into result [#6](https://github.com/Songmu/horenso/pull/6) ([Songmu](https://github.com/Songmu))
* Fix typo [#7](https://github.com/Songmu/horenso/pull/7) ([ariarijp](https://github.com/ariarijp))
* [incompatible] pass the result JSON via STDIN instead of command line argument [#5](https://github.com/Songmu/horenso/pull/5) ([Songmu](https://github.com/Songmu))
* Fix test [#2](https://github.com/Songmu/horenso/pull/2) ([Songmu](https://github.com/Songmu))
* Initial implement [#1](https://github.com/Songmu/horenso/pull/1) ([Songmu](https://github.com/Songmu))
