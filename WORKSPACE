load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "new_git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

#######

#load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
#
#http_archive(
#    name = "rules_python",
#    sha256 = "36362b4d54fcb17342f9071e4c38d63ce83e2e57d7d5599ebdde4670b9760664",
#    strip_prefix = "rules_python-0.18.0",
#    url = "https://github.com/bazelbuild/rules_python/releases/download/0.18.0/rules_python-0.18.0.tar.gz",
#)
#
#load("@rules_python//python:repositories.bzl", "py_repositories")
#
#py_repositories()

########

git_repository(
    name = "platformio_rules",
    remote = "http://github.com/mum4k/platformio_rules.git",
    tag = "v0.0.10",
)

new_git_repository(
    name = "adafruit_bme280",
    remote = "https://github.com/adafruit/Adafruit_BME280_Library.git",
    tag = "2.2.2",
    build_file = "adafruit_bme280.BUILD",
)

new_git_repository(
    name = "u8g2",
    commit = "3500b1056b33999b285387d196c78fe4ab707f79",
    remote = "https://github.com/olikraus/u8g2.git",
    build_file = "u8g2.BUILD",
)

new_git_repository(
    name = "nanopb",
    remote = "https://github.com/nanopb/nanopb.git",
    tag = "nanopb-0.4.7",
    build_file = "nanopb.BUILD",
)

new_git_repository(
    name = "adafruit_seesaw",
    remote = "https://github.com/adafruit/Adafruit_Seesaw.git",
    tag = "1.6.3",
    build_file = "adafruit_seesaw.BUILD",
)
