# Changelog

## [unreleased]((https://github.com/NodeFactoryIo/vedran/tree/HEAD))

[Full Changelog](https://github.com/NodeFactoryIo/vedran/compare/v.0.1.1...HEAD)

### Added
- Implement tunnel server in load balancer [\#63](https://github.com/NodeFactoryIo/vedran/pull/63) ([MakMuftic](https://github.com/MakMuftic))
- Implement penalizing nodes [\#68](https://github.com/NodeFactoryIo/vedran/pull/68) ([MakMuftic](https://github.com/MakMuftic))

### Fix
- Fix panic if no nodes in database after startup [\#62](https://github.com/NodeFactoryIo/vedran/pull/62) ([mpetrun5](https://github.com/mpetrun5))
- Fix error on provided public IP [\#72](https://github.com/NodeFactoryIo/vedran/pull/72) ([MakMuftic](https://github.com/MakMuftic))
- Fix error saving node cooldown [\#82](https://github.com/NodeFactoryIo/vedran/pull/82) ([mpetrun5](https://github.com/mpetrun5))

### Changed
- Use port from tunnel map for calling node [\#65](https://github.com/NodeFactoryIo/vedran/pull/65) ([mpetrun5](https://github.com/mpetrun5))
- Refactor whitelisting [\#75](https://github.com/NodeFactoryIo/vedran/pull/75) ([MakMuftic](https://github.com/MakMuftic))

## [v0.1.1]((https://github.com/NodeFactoryIo/vedran/tree/v0.1.1))

[Full Changelog](https://github.com/NodeFactoryIo/vedran/compare/v.0.1.0...v.0.1.1)

### Added
- Docker compose setup [\#24](https://github.com/NodeFactoryIo/vedran/pull/24) ([MakMuftic](https://github.com/MakMuftic))

### Fix

### Changed

## [v0.1.0]((https://github.com/NodeFactoryIo/vedran/tree/v0.1.0))

[Full Changelog](https://github.com/NodeFactoryIo/vedran/compare/6facfb9564b9da01e3652117334c94774da3360e...v.0.1.0)

### Added
- Initial repo setup [\#9](https://github.com/NodeFactoryIo/vedran/pull/9) ([MakMuftic](https://github.com/MakMuftic))
- Optimize Dockerfile [\#14](https://github.com/NodeFactoryIo/vedran/pull/14) ([MakMuftic](https://github.com/MakMuftic))
- API routes setup [\#10](https://github.com/NodeFactoryIo/vedran/pull/10) ([MakMuftic](https://github.com/MakMuftic))
- Cache dependecies [\#16](https://github.com/NodeFactoryIo/vedran/pull/16) ([MakMuftic](https://github.com/MakMuftic))
- Generate auth token [\#20](https://github.com/NodeFactoryIo/vedran/pull/20) ([MakMuftic](https://github.com/MakMuftic))
- Add bolt db [\#21](https://github.com/NodeFactoryIo/vedran/pull/21) ([MakMuftic](https://github.com/MakMuftic))
- Add ci build matrix [\#22](https://github.com/NodeFactoryIo/vedran/pull/22) ([mpetrunic](https://github.com/mpetrunic))
- CLI command [\#23](https://github.com/NodeFactoryIo/vedran/pull/23) ([MakMuftic](https://github.com/MakMuftic))
- Ping endpoint [\#25](https://github.com/NodeFactoryIo/vedran/pull/25) ([MakMuftic](https://github.com/MakMuftic))
- Metrics endpoint [\#27](https://github.com/NodeFactoryIo/vedran/pull/27) ([MakMuftic](https://github.com/MakMuftic))
- Implement whitelisting logic [\#29](https://github.com/NodeFactoryIo/vedran/pull/29) ([MakMuftic](https://github.com/MakMuftic))
- Setup logging [\#31](https://github.com/NodeFactoryIo/vedran/pull/31) ([MakMuftic](https://github.com/MakMuftic))

### Fix
- Fix metrics endpoint to PUT [\#30](https://github.com/NodeFactoryIo/vedran/pull/30) ([MakMuftic](https://github.com/MakMuftic))
- Fix invalid port validation [\#26](https://github.com/NodeFactoryIo/vedran/pull/26) ([MakMuftic](https://github.com/MakMuftic))
- Fix failing docker build [\#19](https://github.com/NodeFactoryIo/vedran/pull/19) ([MakMuftic](https://github.com/MakMuftic))

### Changed
