# openmcs

An open source server emulator for your favorite online medieval clicking simulator.

## Building

You'll need at least Go 1.19 to work with this project.

You can build an executable for the server simply by running `make` in the top-level directory. This will run unit tests
and build the code into a binary which you can then run.

## Monitoring

The server is instrumented with Prometheus metrics, available at http://localhost:2112/metrics. In addition to standard
Go metrics, the server exposes several of its own metrics to help provide insights into the various processes which
could impact performance.

* `game_state_update_duration_bucket`: a histogram describing how long game state updates take to complete
* `users_online_total`: a gauge for the current, active player count

This project comes with a Docker Compose stack consisting of a Prometheus and Grafana instance with prebuilt dashboards.

You can start the stack by running `make start-monitoring`, and bring it down by running `make stop-monitoring`. 

Once up, you can access the components at the following locations:

* Grafana: http://localhost:3000
* Prometheus: http://localhost:9090

The default Grafana login is `admin`/`admin`. Prometheus will store its metrics data under the `data/prometheus` 
directory, so it's safe to stop and start the stack as necessary without losing data.
