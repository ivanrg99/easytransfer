package com.ivan.easytransfer;

import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;

import java.io.File;
import java.io.IOException;
import java.util.concurrent.ExecutorService;

@RestController
public class FileUploadController {
    private final StringBuilder sb;
    private final ExecutorService threadPool;
    public FileUploadController(ExecutorService executorService) {
        this.sb = new StringBuilder();
        this.threadPool = executorService;
    }

    @PostMapping("/upload")
    public String handleFileUpload(@RequestParam("file") MultipartFile file) throws IOException {
        sb.setLength(0);
        sb.append(System.getProperty("user.home"));
        sb.append("/Downloads/");
        sb.append(file.getOriginalFilename());
        File f = new File(sb.toString());
        file.transferTo(f);
        return "Ok";
    }
}
