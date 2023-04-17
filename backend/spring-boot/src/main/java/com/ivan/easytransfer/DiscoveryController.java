package com.ivan.easytransfer;

import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

@RestController
public class DiscoveryController {
    final ExecutorService threadPool;
    final Worker[] workers;
    final RestTemplate restTemplate;

    ArrayList<Device> devices;
    public DiscoveryController(RestTemplate restTemplate, ExecutorService executorService) {
        // this.threadPool = Executors.newFixedThreadPool(Runtime.getRuntime().availableProcessors());
        this.threadPool = executorService;
        this.workers = new Worker[Runtime.getRuntime().availableProcessors()];
        this.devices = new ArrayList<>(5);
        this.restTemplate = restTemplate;
    }

    // @CrossOrigin(origins = "http://localhost:3287")
    @GetMapping("/devices")
    public List<Device> getNetworkIPs() throws InterruptedException, ExecutionException {
        this.devices.clear();
        int range = 255 / workers.length;
        for (int index = 0; index < this.workers.length; index++) {
            int start = index * range;
            int end = start + range;
            this.workers[index] = new Worker(start, end, this.restTemplate);
        }

        List<Future<ArrayList<Device>>> results = this.threadPool.invokeAll(Arrays.asList(this.workers));
            for (Future<ArrayList<Device>> future : results) {
                this.devices.addAll(future.get());
            }
            return this.devices;
    }

    @GetMapping("/devices_cache")
    public List<Device> getNetworkIPsCached() throws InterruptedException, ExecutionException {
        if (this.devices.size() != 0) {
            return this.devices;
        }
        return getNetworkIPs();
    }

    @GetMapping("/health")
    public String getHealth() {
        return "Healthy!";
    }

    @GetMapping("/device/name")
    public String getDeviceName() throws UnknownHostException {
        return InetAddress.getLocalHost().getHostName();
    }

}
