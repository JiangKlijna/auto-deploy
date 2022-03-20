
// init
(function () {
    var app = document.getElementById("app");
    var main = document.getElementById("main");

    window.onresize = function (ev) {
        if (window.innerWidth < 800 || (window.innerWidth < 1200 && window.innerWidth >= window.innerHeight)) { // mobile style
            main.style.marginTop = 0;
            main.style.width = '100%';
            main.style.height = '100%';
        } else { // pc style
            var marginTop = parseInt((window.innerWidth-800) / (window.screen.width-800) * 100);
            main.style.marginTop = marginTop + 'px';
            main.style.height = (window.innerHeight - marginTop * 2) + 'px';
            main.style.width = '800px';
        }
    };
    window.onresize();

    // random choose or read local
    var random_choose = function (key, arr, isReadLocal) {
        if (isReadLocal) {
            var value = localStorage.getItem("random_choose_"+key);
            if (value) {
                return value;
            }
        }
        var value = arr[parseInt(Math.random()*arr.length)];
        localStorage.setItem("random_choose_"+key, value);
        return value;
    }
    var R = {
        accent_colors: ["amber", "blue", "cyan", "deep-orange", "deep-purple", "green", "indigo", "light-blue", "light-green", "lime", "orange", "pink", "purple", "red", "teal", "yellow"],
        // primary_colors: ["amber", "blue", "blue-grey", "brown", "cyan", "deep-orange", "deep-purple", "green", "grey", "indigo", "light-blue", "light-green", "lime", "orange", "pink", "purple", "red", "teal", "yellow"],
    };
    var change_color = function (e) {
        // var primary_color = random_choose("primary_colors", R.primary_colors, e===undefined);
        var accent_color = random_choose("accent_colors", R.accent_colors, e===undefined);
        app.className = 'mdui-theme-primary-' + accent_color + ' mdui-theme-accent-' + accent_color;
    }
    change_color();
    document.getElementById("change-color").onclick = change_color;
})();
