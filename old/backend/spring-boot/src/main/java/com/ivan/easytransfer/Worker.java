package com.ivan.easytransfer;

import org.springframework.web.client.RestTemplate;

import java.net.InetAddress;
import java.util.ArrayList;
import java.util.concurrent.Callable;

public class Worker implements Callable<ArrayList<Device>> {

    private final int start;
    private final int end;
    RestTemplate restTemplate;

    public Worker(int start, int end, RestTemplate restTemplate) {
        this.start = start;
        this.end = end;
        this.restTemplate = restTemplate;
    }

    @Override
    public ArrayList<Device> call() throws Exception {
        var host = InetAddress.getLocalHost().toString();
        var host_ip = host.substring(host.indexOf('/') + 1);

        var results = new ArrayList<Device>();
        final byte[] ip;
        try {
            ip = InetAddress.getLocalHost().getAddress();
        } catch (Exception e) {
            return results;
        }

        for (int i = start; i < end; i++) {
            ip[3] = (byte) i;
            InetAddress address = InetAddress.getByAddress(ip);
            String output = address.toString().substring(1);

            if (output.equals(host_ip)) {
                continue;
            }
            try {
                StringBuilder sb = new StringBuilder(40);
                sb.append("http://").append(output).append(":").append(Config.PORT());
                sb.append("/health");
                String response = this.restTemplate.getForObject(sb.toString(), String.class);
                if (response.equals("Healthy!")) {
                    sb.setLength(0);
                    sb.append("http://").append(output).append(":").append(Config.PORT());
                    sb.append("/device/name");
                    String name = this.restTemplate.getForObject(sb.toString(), String.class);
                    results.add(new Device(output, name));
                }
            } catch (Exception e) {
            }
        }
        return results;
    }
}
