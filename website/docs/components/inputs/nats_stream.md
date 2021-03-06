---
title: nats_stream
type: input
categories: ["Services"]
---

<!--
     THIS FILE IS AUTOGENERATED!

     To make changes please edit the contents of:
     lib/input/nats_stream.go
-->

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


Subscribe to a NATS Stream subject. Joining a queue is optional and allows
multiple clients of a subject to consume using queue semantics.


<Tabs defaultValue="common" values={[
  { label: 'Common', value: 'common', },
  { label: 'Advanced', value: 'advanced', },
]}>

<TabItem value="common">

```yaml
# Common config fields, showing default values
input:
  nats_stream:
    urls:
      - nats://127.0.0.1:4222
    cluster_id: test-cluster
    client_id: benthos_client
    queue: benthos_queue
    subject: benthos_messages
    durable_name: benthos_offset
    unsubscribe_on_close: false
```

</TabItem>
<TabItem value="advanced">

```yaml
# All config fields, showing default values
input:
  nats_stream:
    urls:
      - nats://127.0.0.1:4222
    cluster_id: test-cluster
    client_id: benthos_client
    queue: benthos_queue
    subject: benthos_messages
    durable_name: benthos_offset
    unsubscribe_on_close: false
    start_from_oldest: true
    max_inflight: 1024
    ack_wait: 30s
```

</TabItem>
</Tabs>

Tracking and persisting offsets through a durable name is also optional and
works with or without a queue. If a durable name is not provided then subjects
are consumed from the most recently published message.

When a consumer closes its connection it unsubscribes, when all consumers of a
durable queue do this the offsets are deleted. In order to avoid this you can
stop the consumers from unsubscribing by setting the field
`unsubscribe_on_close` to `false`.

### Metadata

This input adds the following metadata fields to each message:

``` text
- nats_stream_subject
- nats_stream_sequence
```

You can access these metadata fields using
[function interpolation](/docs/configuration/interpolation#metadata).

## Fields

### `urls`

A list of URLs to connect to. If an item of the list contains commas it will be expanded into multiple URLs.


Type: `array`  
Default: `["nats://127.0.0.1:4222"]`  

```yaml
# Examples

urls:
  - nats://127.0.0.1:4222
```

### `cluster_id`

The ID of the cluster to consume from.


Type: `string`  
Default: `"test-cluster"`  

### `client_id`

A client ID to connect as.


Type: `string`  
Default: `"benthos_client"`  

### `queue`

The queue to consume from.


Type: `string`  
Default: `"benthos_queue"`  

### `subject`

A subject to consume from.


Type: `string`  
Default: `"benthos_messages"`  

### `durable_name`

Preserve the state of your consumer under a durable name.


Type: `string`  
Default: `"benthos_offset"`  

### `unsubscribe_on_close`

Whether the subscription should be destroyed when this client disconnects.


Type: `bool`  
Default: `false`  

### `start_from_oldest`

If a position is not found for a queue, determines whether to consume from the oldest available message, otherwise messages are consumed from the latest.


Type: `bool`  
Default: `true`  

### `max_inflight`

The maximum number of unprocessed messages to fetch at a given time.


Type: `number`  
Default: `1024`  

### `ack_wait`

An optional duration to specify at which a message that is yet to be acked will be automatically retried.


Type: `string`  
Default: `"30s"`  


