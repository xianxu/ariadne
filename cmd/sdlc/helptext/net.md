Read-only observations for an external network incident. Git transport errors
print `infrastructure: network` and the exact command to diagnose their host;
they perform no probes and never retry an uncertain remote write.

    sdlc net check
    sdlc net check --remote origin
    sdlc net check --host github.com

Without --host, use the configured publication remote's effective fetch host.
Distinct push hosts are named with their separate check command. --host works
outside a repository. github.com also checks api.github.com DNS; --api-host
can name another API host. No API endpoint is guessed for other Git hosts.

The system lookup uses dscacheutil on macOS or getent on Linux. The direct
lookup uses dig against its configured default DNS server, bypassing the OS
resolver cache. Missing utilities and unsupported systems report unavailable;
this command installs nothing and does not change DNS or flush caches.

Both DNS paths feed independent HTTPS attempts to the Git host. HTTPS uses one
sampled address per path, retains hostname/certificate validation, follows no
redirects, sends no credentials, and uses no environmental proxy. The output
labels this limit: it does not establish that Git authentication, SSH or its
proxy route works. An HTTP response establishes reachability, not service health.

The whole check has an eight-second budget, at most four concurrent effects,
and bounded utility output. Ctrl+C or SIGTERM cancels and joins the probes.
The report includes partial observations before a nonzero exit on any failure
or unavailable requested observation. A path disagreement suggests a resolver
path difference; it does not prove an upstream outage or a specific cache fault.
Report the incident to the hub; preserve existing operation recovery state.

FLAGS
  --host HOST       explicit hostname or IP; excludes --remote
  --remote NAME     explicit configured Git remote
  --api-host HOST   optional additional DNS-only API observation
