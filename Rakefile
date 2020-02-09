SOURCES_DIR = 'cmd/nswitchcov_a'
DIST_DIR = 'dist'
CMD_NAME = "nswitchcov_a"

task :default => :build
task :all => [:build]

task :build do
  sh "go build #{SOURCES_DIR}/main.go"
end

def create_release_directory(app_name, os_name)
  File.open("#{SOURCES_DIR}/main.go") do |file|
    version = ""
    file.each_line do |line|
      if line.include?('NSwitchCovAVersion')
        md = line.match(/[.0-9]+/)
        if md then
          version = md[0]
          break
        end
      end
    end

    return if version.empty?

    dict_os_name = os_name
    dict_os_name = "macosx" if os_name == "darwin"
    
    folder_name = "#{CMD_NAME}" + "_" + dict_os_name + "_v" + version
    FileUtils.rm_rf("#{DIST_DIR}/#{folder_name}")
    FileUtils.cp_r("#{DIST_DIR}/nswitchcov_a_env_version", "#{DIST_DIR}/#{folder_name}")
    FileUtils.mv(app_name, "#{DIST_DIR}/#{folder_name}/#{app_name}")
  end
end

task :cross_build do
  cmd_name_param = {"windows" => "#{CMD_NAME}.exe", "darwin" => "#{CMD_NAME}"}

  for os in ["windows", "darwin"]
    sh "set GOOS=#{os}&set GOARCH=amd64&go build -o #{cmd_name_param[os]} -ldflags=\"-X 'main.TargetEnv=#{os}'\" #{SOURCES_DIR}/main.go"
    create_release_directory(cmd_name_param[os], os)
  end
end