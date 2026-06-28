cmake_minimum_required(VERSION 3.13)
project(example LANGUAGES C)

set(CMAKE_C_STANDARD 23)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE "Release" CACHE STRING "Choose the type of build." FORCE)
endif()
set_property(CACHE CMAKE_BUILD_TYPE PROPERTY STRINGS "Debug" "Release")

if(CMAKE_BUILD_TYPE STREQUAL "Release")
  set(ENABLE_NDEBUG ON CACHE BOOL "Define NDEBUG globally (disables assert())." FORCE)
  add_compile_definitions(NDEBUG)
  message(STATUS "Global NDEBUG enabled (build type: ${CMAKE_BUILD_TYPE})")
else()
  set(ENABLE_NDEBUG OFF CACHE BOOL "Define NDEBUG globally (disables assert.)." FORCE)
  message(STATUS "Global NDEBUG disabled active (build type: ${CMAKE_BUILD_TYPE})")
endif()

include(GNUInstallDirs)

enable_testing()
add_subdirectory(tests)

if(CMAKE_C_COMPILER_ID MATCHES "GNU|Clang")
  add_compile_options(
        -Wall
        -Wextra
        -Wpedantic
        -Wconversion
        -Wsign-conversion
        -Wshadow
        -Wcast-align
        -Wformat=2
        -Wundef
        -Wdouble-promotion
        -Wnull-dereference

        $<$<CONFIG:Debug>:-O0>
        $<$<CONFIG:Debug>:-g3>
        $<$<CONFIG:Debug>:-ggdb>
        $<$<CONFIG:Debug>:-funsigned-char>

        $<$<CONFIG:Release>:-O3>
        $<$<CONFIG:Release>:-funsigned-char>
    )
else()
  message(FATAL_ERROR, "Compiler not suported :(")
endif()

add_library(lib STATIC
  lib/lib.c
)
target_include_directories(lib PUBLIC
  ${CMAKE_CURRENT_SOURCE_DIR}/lib
)
add_library(example::lib ALIAS lib)
add_executable(example example/main.c)
