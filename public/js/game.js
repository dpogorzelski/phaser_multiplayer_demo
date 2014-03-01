window.onload = function() {
    var game = new Phaser.Game(1024, 768, Phaser.AUTO, null, {
        preload: preload,
        create: create,
        update: update
    });

    function preload() {
        game.load.spritesheet('char', 'images/char01.png', 32, 48);
    }

    var player;
    var players = {};
    var sock;
    var label;
    var style = {
        font: "12px Arial",
        fill: "#ffffff"
    };
    var ip = "192.168.1.13";

    function spawn(m) {
        var label = m.Id.match(/(^\w*)-/i)[1];
        var p = game.add.sprite(m.X, m.Y, 'char');
        p.animations.add('down', [0, 1, 2], 10);
        p.animations.add('left', [12, 13, 14], 10);
        p.animations.add('right', [24, 25, 26], 10);
        p.animations.add('up', [36, 37, 38], 10);
        p.label = game.add.text(m.X, m.Y - 10, label, style);
        return p;
    }

    function uPosition(m) {
        if (players[m.Id].x > m.X) {
            players[m.Id].animations.play('left');
        } else if (players[m.Id].x < m.X) {
            players[m.Id].animations.play('right');
        } else if (players[m.Id].y > m.Y) {
            players[m.Id].animations.play('up');
        } else {
            players[m.Id].animations.play('down');
        }
        players[m.Id].x = players[m.Id].label.x = m.X;
        players[m.Id].y = m.Y;
        players[m.Id].label.y = m.Y - 10;
    }

    function create() {
        game.stage.backgroundColor = '#2d2d2d';

        player = game.add.sprite(0, 0, 'char');

        player.animations.add('down', [0, 1, 2], 10);
        player.animations.add('left', [12, 13, 14], 10);
        player.animations.add('right', [24, 25, 26], 10);
        player.animations.add('up', [36, 37, 38], 10);

        player.body.collideWorldBounds = true;

        sock = new WebSocket("ws://" + ip + ":3000/ws");
        sock.onopen = function() {
            var pos = JSON.stringify({
                x: player.x,
                y: player.y
            });
            sock.send(pos);
        };

        sock.onmessage = function(message) {
            var m = JSON.parse(message.data);
            if (m.New) {
                players[m.Id] = spawn(m);
            } else if (m.Online === false) {
                players[m.Id].label.destroy();
                players[m.Id].destroy();
            } else {
                uPosition(m);
            }
        };
    }

    function update() {
        if (game.input.keyboard.isDown(Phaser.Keyboard.LEFT)) {
            player.animations.play('left');
            player.x -= 3;
            var pos = JSON.stringify({
                x: player.x
            });
            sock.send(pos);
        } else if (game.input.keyboard.isDown(Phaser.Keyboard.RIGHT)) {
            player.animations.play('right');
            player.x += 3;
            var pos = JSON.stringify({
                x: player.x
            });
            sock.send(pos);
        } else if (game.input.keyboard.isDown(Phaser.Keyboard.UP)) {
            player.animations.play('up');
            player.y -= 3;
            var pos = JSON.stringify({
                y: player.y
            });
            sock.send(pos);

        } else if (game.input.keyboard.isDown(Phaser.Keyboard.DOWN)) {
            player.animations.play('down');
            player.y += 3;
            var pos = JSON.stringify({
                y: player.y
            });
            sock.send(pos);
        }
    }
};
