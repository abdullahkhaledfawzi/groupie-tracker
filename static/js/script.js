/**
 * Groupie Tracker — Client-Side JavaScript
 *
 * Implements the required client-server event:
 * Live search that sends requests to the Go server via fetch API,
 * receives JSON responses, and dynamically updates the artist cards.
 *
 * Performance: uses debouncing (300ms) to avoid excessive requests.
 * DOM updates use innerHTML replacement — no layout thrashing.
 */
document.addEventListener("DOMContentLoaded", function () {
    var searchInput = document.getElementById("search-input");
    var artistsGrid = document.getElementById("artists-grid");
    var noResults = document.getElementById("no-results");

    // Guard: only run on the home page where these elements exist.
    if (!searchInput || !artistsGrid || !noResults) {
        return;
    }

    // Store the original HTML so we can restore it when search is cleared.
    var originalCards = artistsGrid.innerHTML;
    var debounceTimer = null;

    /**
     * Event listener: fires on every keystroke in the search input.
     * Triggers a client→server request after a 300ms debounce delay.
     * This is the required "client call to the server" (request-response pattern).
     */
    searchInput.addEventListener("input", function (e) {
        clearTimeout(debounceTimer);

        debounceTimer = setTimeout(function () {
            var query = e.target.value.trim();

            // If query is empty, restore all cards without a server call.
            if (query === "") {
                artistsGrid.innerHTML = originalCards;
                noResults.style.display = "none";
                return;
            }

            // Client→Server: send search request to Go server.
            fetch("/search?q=" + encodeURIComponent(query))
                .then(function (response) {
                    if (!response.ok) {
                        throw new Error("Server responded with status " + response.status);
                    }
                    return response.json();
                })
                .then(function (artists) {
                    // Server→Client: received JSON response, update the DOM.
                    if (artists.length === 0) {
                        artistsGrid.innerHTML = "";
                        noResults.style.display = "block";
                        return;
                    }

                    noResults.style.display = "none";

                    // Build new cards from the search results.
                    var html = "";
                    for (var i = 0; i < artists.length; i++) {
                        var a = artists[i];
                        html +=
                            '<a href="/artist?id=' + a.id + '" class="card" id="artist-card-' + a.id + '">' +
                                '<div class="card-image">' +
                                    '<img src="' + a.image + '" alt="' + a.name + '" loading="lazy" width="300" height="300">' +
                                '</div>' +
                                '<div class="card-info">' +
                                    '<h3 class="card-name">' + a.name + '</h3>' +
                                    '<p class="card-year">Since ' + a.creationDate + '</p>' +
                                '</div>' +
                            '</a>';
                    }
                    artistsGrid.innerHTML = html;
                })
                .catch(function (err) {
                    console.error("Search error:", err);
                });
        }, 300);
    });
});
