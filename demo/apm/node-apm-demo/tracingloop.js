const http = require('http');
var request = require('request');


if (process.env.MELT_AUTOGENERATE_TRACING_DATA) {
    setTimeout(() => {}, 5000);

    setInterval(()=>{

        http.get('http://localhost:3002/api/tutorials');

        http.get('http://localhost:3002/error');

        http.get('http://localhost:3002/500-error');

        http.get('http://localhost:3002/504-error');

        request.post('http://localhost:3002/api/tutorials', {
            "title": "git3",
            "description": "test description3"
        }, (error, response, body) => {
            let body_new = JSON.parse(body);
            request.put(`http://localhost:3002/api/tutorials/${body_new.id}`, {
                "title": "git4",
                "description": "test description4"
                }, (puterror, putresponse, putbody) => {

                    request.delete(`http://localhost:3002/api/tutorials/${body_new.id}`);
            });
        });

    },3000);
}