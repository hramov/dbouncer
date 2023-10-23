const net = require('net');

let appId = null;
const client = new net.Socket();

client.on('close', function() {
    console.log('Connection closed');
});

client.on('data', function(d) {
    const obj = JSON.parse(d.toString());
    appId = obj.app_id;

    const data = [];

    for (const r of obj.result) {
        if (r.length === 1) continue;
        data.push(JSON.parse(r))
    }

    console.log(data);
});

client.connect(2379, '127.0.0.1', function() {
    const obj = {
        id:       1,
        app_id:    appId,
        app_name: 'test',
        database: 'postgres',
        kind:     'query',
        query:    'select * from pg_stat_activity',
        params:   null,
    }

    // for (let i = 0; i < 10; i++) {
        client.write(JSON.stringify(obj));
    // }
});