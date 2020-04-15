# MusicAppGo
Music App with backend written in go and backend divided in multiple services

With this application users can discover new music. Users can like and dislike songs. They will get suggestion of songs which they may like. These suggestions are based on the prefences of the user. This is an ordered list based on preferences of other users where the weight of the preferences of a particular user is the amount of similarities with the original user. Furthermore, users can follow artists. These preferences will be used to notify the user when an artist released new songs. An admin can add a new artist to the database or can add the spotify id to an existing artist. The database will be regularly updated by sending requests to an Spotify endpoint.

TO Do:
    - Adding new songs to the music database
    - Adding new albums to the music database
    - Adding extra tags to songs or albums
    - Adding handlers to find songs of a given artist or to search the database
    - Adding handlers for likes, dislikes and follow
    - Adding suggestion system
    - Adding update via Spotify
    - Adding events
    - Create configuration files
    - Creating Docker Images of every service
    - Adding front-end
