file(GLOB TEST_SOURCES CONFIGURE_DEPENDS
    "${CMAKE_CURRENT_SOURCE_DIR}/src/*.c"
)

foreach(test_src IN LISTS TEST_SOURCES)
  get_filename_component(raw_name "${test_src}" NAME_WE)


  set(test_target "test_${raw_name}")

  add_executable(${test_target} "${test_src}")

  target_include_directories(${test_target} PRIVATE
        ${PROJECT_SOURCE_DIR}/include
    )
  target_link_libraries(${test_target} PRIVATE
        example::lib
    )


  add_test(NAME ${test_target} COMMAND ${test_target})

  message(STATUS "Added test target: ${test_target}")
endforeach()
