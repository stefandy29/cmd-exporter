# cmd-exporter
cmd-exporter is a tool designed for Linux that collects data based on linux command, then it will converted to metric.

## Usage
```
./cmd-exporter --config.file=server.yaml
```

<h3 id="Config-Usage">Config Usage</h3>
<table>
<thead>
<tr>
<th>Name</th>
<th>Description</th>
<th>Value Example</th>
<th>Required</th>
</tr>
</thead>
<tbody><tr>
<td>Port</td>
<td>Metric port address</td>
<td>9126</td>
<td>Yes</td>
</tr>
<tr>
<td>timeout_server</td>
<td>Maximum duration for reading the entire request, including the body. A zero or negative value means there will be no timeout.</td>
<td>5</td>
<td>Yes</td>
</tr>
<tr>
<td>username</td>
<td>Username for basic auth</td>
<td>root</td>
<td>No</td>
</tr>
<tr>
<td>password</td>
<td>Password for basic auth</td>
<td>password</td>
<td>No</td>
</tr>
<tr>
<td>certfile</td>
<td>Certificate file for TLS connection</td>
<td>server.crt</td>
<td>No</td>
</tr>
<tr>
<td>keyfile</td>
<td>Key file for TLS connection</td>
<td>server.key</td>
<td>No</td>
</tr>
<tr>
<td>process_name</td>
<td>Prometheus metric name</td>
<td>cpu_usage_total</td>
<td>Yes</td>
</tr>
<tr>
<td>command</td>
<td>Linux command for producing a value for metric</td>
<td>"grep 'cpu ' /proc/stat | awk '{usage=($2+$4)*100/($2+$4+$5)} END {print usage }' |  xargs printf '%.*f\\n' '2'"</td>
<td>Yes</td>
</tr>
</tbody></table>