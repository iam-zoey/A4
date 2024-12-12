Start the Simulation

To start the simulation, use the following command:
```
./run.sh
```
This starts all necessary nodes (Traders and Sellers) as separate processes.
In the `/log` , you can find the log of each process, indicating how they handle seller requests and monitor themselves with a heartbeat protocol


To simulate a Trader failure, use the following command to kill Trader 1:
```
./kill-trader1.sh
```
For macOS users, if the above script doesn’t work, use:
```
pkill -f "trader.go -id=1"
```

This simulates a failure in Trader 1, allowing Trader 2 to take over leadership. As soon as seller lost connection with trader, it will restart requests until the other trader picks up and respond.


To stop all nodes (Traders and Sellers) and cleanly exit the simulation, use:
```
./finish-simulation.sh
```