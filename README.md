<h1>What It Does</h1>
<p>This app allows you to automatically download megascans content which you have already purchased into a folder of your choosing</p>
<h1>Downloading App</h1>
<p>Your app may get flagged by your browser and windows defender as being a Trojan horse virus. To get around this, you will have to enable the 
app in windows defender once it is downloaded, otherwise windows will try to delete it.
</p>

<p>Once downloaded, the app can be run standalone or executed in a terminal -- the terminal is the preffered way of running it.</p>

<h1>Choosing an install location</h1>
<p>You will next be asked to add a download location, put it somehwere which will have enough space for your purchased megascans assets.
</p>

<h1>Authentication</h1>
You will then be asked to authenticate
<ul>
  <li>Enter your email associated with your megascans account.</li>
  <li>Enter Bearer auth token from browser (see below on how to do this)</li>
</ul>

<h1>Getting a bearer auth token</h1>
<ol>
  <li> Sign In To Megascans Account on https://quixel.com/megascans/home </li>
  <li>
    Go to Megascans Account Page https://quixel.com/account</a>
  </li>
  <li>Open Developer Tools (cntrl + shift + i) and click on Network Tab and then refresh page</li>
  <li>Look under the network request files for a file matching your email address</li>
  <li>Click on this file, go to the the headers and copy the Authorization token. DO NOT INCLUDE THE "Bearer " part of the token, just everything else </li>
  <li>Paste the bearer token into the app as your login.</li>
  <li>Authorization tokens expire, so every now and then the app will tell you to get a new one following the same process listed above.</li>
</ol>
![authTHing](https://github.com/user-attachments/assets/008795d3-2245-4ed1-a1c5-7403e6b921ec)



